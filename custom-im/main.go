// im-select-win: Windows input method switcher (single-file)
// Build: go build -o im-select-win.exe main.go
//
// 功能：
// - ./im-select-win            -> show current input method (no help by default)
// - ./im-select-win --help     -> show help
// - ./im-select-win --list     -> list available keyboard layouts & TSF profiles
// - ./im-select-win <arg>      -> switch to layout/profile by KLID|HKL|LANGID|NAME_SUBSTR
//
// Implementation notes:
// - Uses user32 LoadKeyboardLayout/ActivateKeyboardLayout + WM_INPUTLANGCHANGEREQUEST
// - If classic path fails (TSF input methods), uses msctf TF_CreateInputProcessorProfiles
//   and ITfInputProcessorProfileMgr::EnumProfiles + ActivateProfile to activate TSF profiles.
// - Carefully handles 64-bit HKL values and avoids truncation bugs.

package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

//
// --- Low-level Windows constants & helpers ---
//

const (
	WM_INPUTLANGCHANGEREQUEST = 0x0050
	KLF_ACTIVATE              = 0x00000001
	// TF profile types (from msctf.h)
	TF_PROFILETYPE_INPUTPROCESSOR = 0x0001
	TF_PROFILETYPE_KEYBOARDLAYOUT = 0x0002
)

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// Useful known GUIDs / CLSIDs / IIDs used by TSF (values from msctf.h and docs)
var (
	CLSID_TF_InputProcessorProfiles = GUID{0x33C53A50, 0xF456, 0x4884, [8]byte{0xB0, 0x49, 0x85, 0xFD, 0x64, 0x3E, 0xCF, 0xED}}
	// IID for ITfInputProcessorProfileMgr
	IID_ITfInputProcessorProfileMgr = GUID{0x71C6E74C, 0x0F28, 0x11D8, [8]byte{0xA8, 0x2A, 0x00, 0x06, 0x5B, 0x84, 0x43, 0x5C}}
	// Category: keyboard TIP (for GetActiveProfile)
	GUID_TFCAT_TIP_KEYBOARD = GUID{0x34745C63, 0xB2F0, 0x4784, [8]byte{0x8B, 0x67, 0x5E, 0x12, 0xC8, 0x70, 0x1A, 0x31}}
)

// TF_INPUTPROCESSORPROFILE (C struct layout matched; careful about alignment)
type TF_INPUTPROCESSORPROFILE struct {
	DwProfileType uint32
	Langid        uint16
	_pad          uint16 // padding
	Clsid         GUID
	GuidProfile   GUID
	Catid         GUID
	HklSubstitute uintptr
	DwCaps        uint32
	Hkl           uintptr
	DwFlags       uint32
}

//
// --- WinAPI procs we use ---
//

var (
	user32                              = syscall.NewLazyDLL("user32.dll")
	kernel32                            = syscall.NewLazyDLL("kernel32.dll")
	msctf                               = syscall.NewLazyDLL("msctf.dll")
	procGetForegroundWindow             = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId        = user32.NewProc("GetWindowThreadProcessId")
	procAttachThreadInput               = user32.NewProc("AttachThreadInput")
	procGetCurrentThreadId              = kernel32.NewProc("GetCurrentThreadId")
	procLoadKeyboardLayoutW             = user32.NewProc("LoadKeyboardLayoutW")
	procActivateKeyboardLayout          = user32.NewProc("ActivateKeyboardLayout")
	procSendMessageW                    = user32.NewProc("SendMessageW")
	procGetKeyboardLayoutList           = user32.NewProc("GetKeyboardLayoutList")
	procGetKeyboardLayoutNameW          = user32.NewProc("GetKeyboardLayoutNameW")
	procTF_CreateInputProcessorProfiles = msctf.NewProc("TF_CreateInputProcessorProfiles")
)

//
// --- Utility: string <-> UTF16 ptr ---
//

func utf16PtrFromString(s string) *uint16 {
	if s == "" {
		return nil
	}
	u := syscall.StringToUTF16(s)
	return &u[0]
}

//
// --- Registry: keyboard layouts (KLID -> display name) ---
//

type LayoutInfo struct {
	KLID string // subkey name, e.g. "00000409" or "E00E0804"
	Name string // Layout Text (best-effort)
}

func readRegisteredKeyboardLayouts() (map[string]LayoutInfo, error) {
	out := make(map[string]LayoutInfo)
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Keyboard Layouts`, registry.READ)
	if err != nil {
		return out, err
	}
	defer key.Close()
	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return out, err
	}
	for _, n := range names {
		sub, err := registry.OpenKey(key, n, registry.READ)
		if err != nil {
			continue
		}
		val, _, err := sub.GetStringValue("Layout Text")
		if err != nil {
			// Try "Layout Display Name" (can be an indirect resource string like "@...,-123")
			val2, _, err2 := sub.GetStringValue("Layout Display Name")
			if err2 == nil {
				val = val2
			} else {
				val = ""
			}
		}
		sub.Close()
		out[strings.ToUpper(n)] = LayoutInfo{KLID: strings.ToUpper(n), Name: val}
	}
	return out, nil
}

//
// --- Helpers to format HKL/KLID ---
//

// format KLID (8 hex uppercase) from HKL uintptr (use lower 32-bit as KLID is 8 hex digits)
func klidFromHKL(h uintptr) string {
	low32 := uint64(uint32(h)) // only low 32 bits are meaningful for KLID formatting
	return fmt.Sprintf("%08X", low32)
}

// parse KLID string like "E00E0804" or "00000409" (hex) -> uint64
func parseKLIDString(s string) (uint64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	// accept 8-hex or shorter
	if len(s) == 0 {
		return 0, errors.New("empty klid")
	}
	if len(s) > 8 {
		// take low 8 characters (the typical KLID is 8 hex digits)
		s = s[len(s)-8:]
	}
	// pad left
	if len(s) < 8 {
		s = strings.Repeat("0", 8-len(s)) + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	// big-endian bytes -> decode as hex directly
	val := uint64(0)
	for _, c := range b {
		val = (val << 8) | uint64(c)
	}
	return val, nil
}

//
// --- Classic path: LoadKeyboardLayout / Activate + WM message ---
//

func loadKeyboardLayoutByKLID(klid string) (uintptr, error) {
	ptr := utf16PtrFromString(klid)
	ret, _, err := procLoadKeyboardLayoutW.Call(uintptr(unsafe.Pointer(ptr)), uintptr(KLF_ACTIVATE))
	if ret == 0 {
		return 0, fmt.Errorf("LoadKeyboardLayoutW failed: %v", err)
	}
	return uintptr(ret), nil
}

func activateKeyboardLayout(hkl uintptr) (uintptr, error) {
	ret, _, err := procActivateKeyboardLayout.Call(hkl, uintptr(KLF_ACTIVATE))
	if ret == 0 {
		return 0, fmt.Errorf("ActivateKeyboardLayout failed: %v", err)
	}
	return uintptr(ret), nil
}

func sendInputLangChangeToForeground(hkl uintptr) error {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		// fallback: broadcast
		// but we use HWND_BROADCAST = 0xffff
		hwnd = 0xffff
	}
	// Attach thread input so ActivateKeyboardLayout affects correct thread
	var fgThread, curThread uintptr
	r1, _, _ := procGetWindowThreadProcessId.Call(hwnd, 0)
	fgThread = uintptr(r1)
	r2, _, _ := procGetCurrentThreadId.Call()
	curThread = uintptr(r2)
	if fgThread != curThread {
		// attach
		procAttachThreadInput.Call(fgThread, curThread, 1)
		defer procAttachThreadInput.Call(fgThread, curThread, 0)
	}
	// Post/Send WM_INPUTLANGCHANGEREQUEST
	// Use SendMessageW to ensure immediate
	_, _, err := procSendMessageW.Call(hwnd,
		uintptr(WM_INPUTLANGCHANGEREQUEST),
		uintptr(0),
		hkl)
	if err != syscall.Errno(0) {
		// non-fatal
	}
	return nil
}

//
// --- TSF path: TF_CreateInputProcessorProfiles -> ITfInputProcessorProfileMgr -> EnumProfiles / ActivateProfile ---
//
// We'll call TF_CreateInputProcessorProfiles to get an object pointer (IUnknown-derived).
// Then QueryInterface for IID_ITfInputProcessorProfileMgr and call its methods via vtable.
//
// The code below performs raw vtable calls using syscall.SyscallN style (Syscall6/Syscall9).
//

func TF_CreateProfiles() (uintptr, error) {
	var p uintptr
	hr, _, _ := procTF_CreateInputProcessorProfiles.Call(uintptr(unsafe.Pointer(&p)))
	if hr != 0 || p == 0 {
		return 0, fmt.Errorf("TF_CreateInputProcessorProfiles failed: HRESULT=0x%08X", uint32(hr))
	}
	return p, nil
}

// QueryInterface on a COM object (raw pointer) to request given IID (GUID). Returns new interface pointer (out pointer).
func comQueryInterface(obj uintptr, iid *GUID) (uintptr, error) {
	if obj == 0 {
		return 0, errors.New("nil object")
	}
	// vtable pointer
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(obj)) // pointer to vtable
	// QueryInterface is vtbl[0]
	qifn := *(*uintptr)(unsafe.Pointer(vtbl + 0*ptrSize))
	var out uintptr
	// QueryInterface signature: HRESULT QueryInterface(this, REFIID, void **ppv)
	// Use Syscall with 3 args
	hr, _, _ := syscall.Syscall(qifn, 3, obj, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&out)))
	if hr != 0 {
		return 0, fmt.Errorf("QueryInterface failed: HRESULT=0x%08X", uint32(hr))
	}
	return out, nil
}

func comRelease(obj uintptr) {
	if obj == 0 {
		return
	}
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(obj))
	releaseFn := *(*uintptr)(unsafe.Pointer(vtbl + 2*ptrSize)) // Release is vtbl[2]
	syscall.Syscall(releaseFn, 1, obj, 0, 0)
}

// EnumProfiles on ITfInputProcessorProfileMgr: returns enumerator pointer
func enumProfiles(profileMgr uintptr, langid uint16) (uintptr, error) {
	if profileMgr == 0 {
		return 0, errors.New("null profileMgr")
	}
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(profileMgr))
	// EnumProfiles is vtbl index 6 (0:QI 1:AddRef 2:Release 3:ActivateProfile 4:DeactivateProfile 5:GetProfile 6:EnumProfiles)
	enumFn := *(*uintptr)(unsafe.Pointer(vtbl + 6*ptrSize))
	var enumPtr uintptr
	hr, _, _ := syscall.Syscall(enumFn, 3, profileMgr, uintptr(langid), uintptr(unsafe.Pointer(&enumPtr)))
	if hr != 0 || enumPtr == 0 {
		return 0, fmt.Errorf("EnumProfiles failed: HRESULT=0x%08X", uint32(hr))
	}
	return enumPtr, nil
}

// IEnumTfInputProcessorProfiles.Next to fetch 1 profile at a time
func enumNextProfile(enumPtr uintptr) (*TF_INPUTPROCESSORPROFILE, error) {
	if enumPtr == 0 {
		return nil, errors.New("null enum")
	}
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(enumPtr))
	// Next is vtable index 4 for enumerator (0:QI,1:AddRef,2:Release,3:Clone,4:Next,5:Reset,6:Skip)
	nextFn := *(*uintptr)(unsafe.Pointer(vtbl + 4*ptrSize))
	var prof TF_INPUTPROCESSORPROFILE
	var fetched uint32
	// Next(this, ulCount, pProfile, pcFetch)
	hr, _, _ := syscall.Syscall6(nextFn, 4, enumPtr, uintptr(1), uintptr(unsafe.Pointer(&prof)), uintptr(unsafe.Pointer(&fetched)), 0, 0)
	if hr != 0 {
		return nil, fmt.Errorf("Next failed: HRESULT=0x%08X", uint32(hr))
	}
	if fetched == 0 {
		return nil, nil // end
	}
	return &prof, nil
}

func enumRelease(enumPtr uintptr) {
	if enumPtr == 0 {
		return
	}
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(enumPtr))
	releaseFn := *(*uintptr)(unsafe.Pointer(vtbl + 2*ptrSize))
	syscall.Syscall(releaseFn, 1, enumPtr, 0, 0)
}

// ActivateProfile on profileMgr
func activateProfile(profileMgr uintptr, profile *TF_INPUTPROCESSORPROFILE) error {
	if profileMgr == 0 || profile == nil {
		return errors.New("invalid args to activateProfile")
	}
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(profileMgr))
	// ActivateProfile is vtbl index 3
	activateFn := *(*uintptr)(unsafe.Pointer(vtbl + 3*ptrSize))
	// Signature: HRESULT ActivateProfile(this, DWORD dwProfileType, LANGID langid, REFCLSID clsid, REFGUID guidProfile, HKL hkl, DWORD dwFlags)
	// We'll pass clsid & guidProfile pointers
	hr, _, _ := syscall.Syscall9(activateFn, 8,
		profileMgr,
		uintptr(profile.DwProfileType),
		uintptr(profile.Langid),
		uintptr(unsafe.Pointer(&profile.Clsid)),
		uintptr(unsafe.Pointer(&profile.GuidProfile)),
		uintptr(profile.Hkl), // try actual hkl
		uintptr(0),           // dwFlags
		0, 0)
	if hr != 0 {
		// If activation with profile.Hkl failed, try with hklSubstitute (some TSF profiles use substitute HKL)
		hr2, _, _ := syscall.Syscall9(activateFn, 8,
			profileMgr,
			uintptr(profile.DwProfileType),
			uintptr(profile.Langid),
			uintptr(unsafe.Pointer(&profile.Clsid)),
			uintptr(unsafe.Pointer(&profile.GuidProfile)),
			uintptr(profile.HklSubstitute),
			uintptr(0),
			0, 0)
		if hr2 != 0 {
			return fmt.Errorf("ActivateProfile failed: HRESULT=0x%08X / 0x%08X", uint32(hr), uint32(hr2))
		}
	}
	return nil
}

//
// --- High-level: list available items (registry + TSF profiles) ---
//

type Candidate struct {
	Source   string // "KLID" or "TSF"
	KLID     string
	HKL      uintptr
	Langid   uint16
	TypeDesc string // "keyboard layout" or "text service"
	Clsid    *GUID
	Guid     *GUID
	Display  string
}

func listCandidates() ([]Candidate, error) {
	var out []Candidate
	// registry keyboard layouts (KLID -> name)
	regs, _ := readRegisteredKeyboardLayouts()
	for k, info := range regs {
		// For keyboard layouts, KLID is key, attempt to parse into numeric HKL
		klid := k
		val, _ := parseKLIDString(klid)
		out = append(out, Candidate{
			Source:   "KLID",
			KLID:     klid,
			HKL:      uintptr(val),
			Langid:   uint16(val & 0xFFFF),
			TypeDesc: "Keyboard Layout (registry)",
			Display:  info.Name,
		})
	}

	// TSF profiles
	profilesObj, err := TF_CreateProfiles()
	if err == nil && profilesObj != 0 {
		// QueryInterface for profileMgr
		pm, err := comQueryInterface(profilesObj, &IID_ITfInputProcessorProfileMgr)
		if err == nil && pm != 0 {
			// enumerate for LANGIDs we care about - we will enumerate for 0..65535? that's too heavy.
			// Instead enumerate for LANGIDs seen in registry plus a few common ones (0x0404 zh-TW, 0x0804 zh-CN, 0x0409 en-US, etc)
			seenLangs := map[uint16]bool{}
			for _, c := range out {
				seenLangs[c.Langid] = true
			}
			// add common LANGIDs to ensure capture
			common := []uint16{0x0409, 0x0804, 0x0404, 0x0411, 0x0000}
			for _, v := range common {
				seenLangs[v] = true
			}
			for lang := range seenLangs {
				enumPtr, err := enumProfiles(pm, lang)
				if err != nil || enumPtr == 0 {
					continue
				}
				for {
					prof, err := enumNextProfile(enumPtr)
					if err != nil {
						break
					}
					if prof == nil {
						break
					}
					// Determine KLID from hkl or hklSubstitute if present
					var klid string
					if prof.Hkl != 0 {
						klid = klidFromHKL(prof.Hkl)
					} else if prof.HklSubstitute != 0 {
						klid = klidFromHKL(prof.HklSubstitute)
					} else {
						klid = ""
					}
					desc := "Text Service"
					if prof.DwProfileType == TF_PROFILETYPE_KEYBOARDLAYOUT {
						desc = "Keyboard Layout"
					}
					display := fmt.Sprintf("%s (clsid=%08X... profile=%08X...)", desc, prof.Clsid.Data1, prof.GuidProfile.Data1)
					out = append(out, Candidate{
						Source:   "TSF",
						KLID:     klid,
						HKL:      prof.Hkl,
						Langid:   prof.Langid,
						TypeDesc: desc,
						Clsid:    &prof.Clsid,
						Guid:     &prof.GuidProfile,
						Display:  display,
					})
				}
				enumRelease(enumPtr)
			}
			comRelease(pm)
		}
		// release base profilesObj
		comRelease(profilesObj)
	}
	return out, nil
}

//
// --- Status (no-args): show current active input method ---
//

func getActiveTSFProfile() (*TF_INPUTPROCESSORPROFILE, error) {
	profilesObj, err := TF_CreateProfiles()
	if err != nil || profilesObj == 0 {
		return nil, fmt.Errorf("TF_CreateInputProcessorProfiles failed: %v", err)
	}
	defer comRelease(profilesObj)
	pm, err := comQueryInterface(profilesObj, &IID_ITfInputProcessorProfileMgr)
	if err != nil || pm == 0 {
		return nil, fmt.Errorf("QueryInterface ITfInputProcessorProfileMgr failed: %v", err)
	}
	defer comRelease(pm)
	// call GetActiveProfile (vtbl index 10)
	ptrSize := uintptr(unsafe.Sizeof(uintptr(0)))
	vtbl := *(*uintptr)(unsafe.Pointer(pm))
	getActiveFn := *(*uintptr)(unsafe.Pointer(vtbl + 10*ptrSize))
	var profile TF_INPUTPROCESSORPROFILE
	hr, _, _ := syscall.Syscall(getActiveFn, 3, pm, uintptr(unsafe.Pointer(&GUID_TFCAT_TIP_KEYBOARD)), uintptr(unsafe.Pointer(&profile)))
	if hr != 0 {
		return nil, fmt.Errorf("GetActiveProfile failed: HRESULT=0x%08X", uint32(hr))
	}
	return &profile, nil
}

func getActiveKLIDClassic() (string, error) {
	// GetKeyboardLayoutNameW gets the name for current thread
	buf := make([]uint16, 9)
	ret, _, err := procGetKeyboardLayoutNameW.Call(uintptr(unsafe.Pointer(&buf[0])))
	if ret == 0 {
		return "", fmt.Errorf("GetKeyboardLayoutNameW failed: %v", err)
	}
	// convert to string
	s := syscall.UTF16ToString(buf)
	return strings.ToUpper(strings.TrimSpace(s)), nil
}

//
// --- Switch logic: accept different user inputs and try to switch ---
//

func switchByKLID(klid string) error {
	klid = strings.ToUpper(strings.TrimSpace(klid))
	// Try classic LoadKeyboardLayout first (works for most keyboard layouts)
	h, err := loadKeyboardLayoutByKLID(klid)
	if err == nil && h != 0 {
		// Activate + post WM to foreground
		_, _ = activateKeyboardLayout(h)
		_ = sendInputLangChangeToForeground(h)
		// small pause to let system apply
		time.Sleep(60 * time.Millisecond)
		// verify via TSF ActiveProfile or GetKeyboardLayoutName
		// Try TSF active profile
		prof, perr := getActiveTSFProfile()
		if perr == nil && prof != nil {
			// match klid with prof.hkl/prof.hklSubstitute
			if prof.Hkl != 0 && strings.EqualFold(klidFromHKL(prof.Hkl), klid) {
				return nil
			}
			if prof.HklSubstitute != 0 && strings.EqualFold(klidFromHKL(prof.HklSubstitute), klid) {
				return nil
			}
		}
		// if TSF check didn't match, check classic GetKeyboardLayoutName result
		name, _ := getActiveKLIDClassic()
		if name != "" && strings.EqualFold(name, klid) {
			return nil
		}
		// classic path didn't appear to take effect — fallthrough to TSF path
	}

	// TSF path: create profiles object and enumerate to find profile whose hkl/hklSubstitute or language matches requested klid
	targetVal, perr := parseKLIDString(klid)
	if perr != nil {
		return fmt.Errorf("invalid KLID '%s': %v", klid, perr)
	}
	profilesObj, err := TF_CreateProfiles()
	if err != nil || profilesObj == 0 {
		return fmt.Errorf("TF_CreateInputProcessorProfiles failed: %v", err)
	}
	defer comRelease(profilesObj)
	pm, err := comQueryInterface(profilesObj, &IID_ITfInputProcessorProfileMgr)
	if err != nil {
		return fmt.Errorf("QueryInterface ITfInputProcessorProfileMgr failed: %v", err)
	}
	defer comRelease(pm)
	langid := uint16(targetVal & 0xFFFF)
	enumPtr, err := enumProfiles(pm, langid)
	if err != nil || enumPtr == 0 {
		// try enumerating with langid 0 (all)
		enumPtr, err = enumProfiles(pm, 0)
		if err != nil || enumPtr == 0 {
			return fmt.Errorf("EnumProfiles failed for langid %04X and 0", langid)
		}
	}
	defer enumRelease(enumPtr)
	for {
		prof, err := enumNextProfile(enumPtr)
		if err != nil {
			break
		}
		if prof == nil {
			break
		}
		// compare candidate HKL with target
		if prof.Hkl != 0 && uintptr(uint64(prof.Hkl)&0xFFFFFFFF) == uintptr(targetVal) {
			// activate
			if err := activateProfile(pm, prof); err == nil {
				// success: also send WM
				_ = sendInputLangChangeToForeground(prof.Hkl)
				return nil
			}
		}
		if prof.HklSubstitute != 0 && uintptr(uint64(prof.HklSubstitute)&0xFFFFFFFF) == uintptr(targetVal) {
			if err := activateProfile(pm, prof); err == nil {
				_ = sendInputLangChangeToForeground(prof.HklSubstitute)
				return nil
			}
		}
	}
	return fmt.Errorf("no TSF profile found matching KLID %s", klid)
}

func switchByHKLStr(hklStr string) error {
	// Accept formats like "0xE00E0804" or "E00E0804" or decimal
	s := strings.TrimSpace(hklStr)
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		s = s[2:]
	}
	// if hex letters present -> treat as hex
	if strings.ContainsAny(s, "ABCDEFabcdef") {
		// parse hex
		val, err := strconv.ParseUint(s, 16, 64)
		if err != nil {
			return err
		}
		klid := fmt.Sprintf("%08X", uint32(val))
		return switchByKLID(klid)
	}
	// else try decimal
	if v, err := strconv.ParseUint(s, 10, 64); err == nil {
		// treat v as numeric HKL -> format KLID from low 32 bits
		klid := fmt.Sprintf("%08X", uint32(v))
		return switchByKLID(klid)
	}
	// fallback treat as KLID
	return switchByKLID(s)
}

func switchByLangIDStr(langStr string) error {
	s := strings.TrimSpace(langStr)
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		s = s[2:]
	}
	val, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return err
	}
	langid := uint16(val)
	// Attempt: enumerate TSF profiles for this langid and activate the first keyboard profile or text service
	profilesObj, err := TF_CreateProfiles()
	if err != nil || profilesObj == 0 {
		return fmt.Errorf("TF_CreateInputProcessorProfiles failed: %v", err)
	}
	defer comRelease(profilesObj)
	pm, err := comQueryInterface(profilesObj, &IID_ITfInputProcessorProfileMgr)
	if err != nil {
		return fmt.Errorf("QueryInterface ITfInputProcessorProfileMgr failed: %v", err)
	}
	defer comRelease(pm)
	enumPtr, err := enumProfiles(pm, langid)
	if err != nil || enumPtr == 0 {
		return fmt.Errorf("EnumProfiles failed for langid %04X", langid)
	}
	defer enumRelease(enumPtr)
	for {
		prof, err := enumNextProfile(enumPtr)
		if err != nil || prof == nil {
			break
		}
		// prefer keyboard layout type
		if prof.DwProfileType == TF_PROFILETYPE_KEYBOARDLAYOUT || prof.HklSubstitute != 0 || prof.Hkl != 0 {
			if err := activateProfile(pm, prof); err == nil {
				_ = sendInputLangChangeToForeground(prof.Hkl)
				return nil
			}
		}
	}
	return fmt.Errorf("no TSF profile activated for LANGID %04X", langid)
}

func switchByNameSubstr(substr string) error {
	substr = strings.ToLower(strings.TrimSpace(substr))
	// search registry layouts
	regs, _ := readRegisteredKeyboardLayouts()
	for klid, info := range regs {
		if strings.Contains(strings.ToLower(info.Name), substr) || strings.Contains(strings.ToLower(klid), substr) {
			// try this KLID
			if err := switchByKLID(klid); err == nil {
				return nil
			}
		}
	}
	// try TSF profiles (search description strings)
	profilesObj, err := TF_CreateProfiles()
	if err == nil && profilesObj != 0 {
		defer comRelease(profilesObj)
		pm, err := comQueryInterface(profilesObj, &IID_ITfInputProcessorProfileMgr)
		if err == nil && pm != 0 {
			defer comRelease(pm)
			// attempt enumerating common languages
			commonLangs := []uint16{0x0804, 0x0409, 0x0404, 0x0000}
			for _, lang := range commonLangs {
				enumPtr, err := enumProfiles(pm, lang)
				if err != nil || enumPtr == 0 {
					continue
				}
				for {
					prof, err := enumNextProfile(enumPtr)
					if err != nil || prof == nil {
						break
					}
					// Build description: try to get display-friendly description from GUID/clsid fields
					desc := fmt.Sprintf("clsid=%08X profile=%08X", prof.Clsid.Data1, prof.GuidProfile.Data1)
					if strings.Contains(strings.ToLower(desc), substr) {
						if err := activateProfile(pm, prof); err == nil {
							_ = sendInputLangChangeToForeground(prof.Hkl)
							enumRelease(enumPtr)
							return nil
						}
					}
				}
				enumRelease(enumPtr)
			}
		}
	}
	return fmt.Errorf("no layout/profile found matching substring '%s'", substr)
}

//
// --- CLI / main ---
//

func printHelp() {
	fmt.Println(`im-select-win

Usage:
  im-select-win                 Show current input method (status)
  im-select-win --help          Show this help
  im-select-win --list          List available keyboard layouts and TSF profiles
  im-select-win <KLID|HKL|LANGID|NAME_SUBSTR>
                                Switch to the target. Examples:
                                  im-select-win 00000409
                                  im-select-win E00E0804
                                  im-select-win 0xE00E0804
                                  im-select-win 2052
                                  im-select-win "Microsoft Pinyin"
Notes:
 - KLID is an 8-hex string (e.g. 00000409, E00E0804)
 - HKL may be passed as hex (0x...) or decimal
 - The tool will attempt the classic LoadKeyboardLayout path first; if that doesn't
   take effect (TSF/modern IME), it will use the Text Services Framework (TSF)
   profile activation path to activate the proper text service/profile.
`)
}

func printStatus() {
	fmt.Println("Current input method info:")
	// Try TSF GetActiveProfile first
	prof, err := getActiveTSFProfile()
	if err == nil && prof != nil {
		fmt.Printf(" - TSF active profile type: %s\n", func() string {
			if prof.DwProfileType == TF_PROFILETYPE_KEYBOARDLAYOUT {
				return "Keyboard Layout"
			}
			return "Text Service"
		}())
		if prof.Hkl != 0 {
			fmt.Printf(" - HLK (hkl): 0x%016X\n", uint64(prof.Hkl))
			fmt.Printf(" - KLID (from hkl): %s\n", klidFromHKL(prof.Hkl))
		} else if prof.HklSubstitute != 0 {
			fmt.Printf(" - HLK (hklSubstitute): 0x%016X\n", uint64(prof.HklSubstitute))
			fmt.Printf(" - KLID (from hklSubstitute): %s\n", klidFromHKL(prof.HklSubstitute))
		}
		fmt.Printf(" - LangID: 0x%04X\n", prof.Langid)
		fmt.Printf(" - CLSID: %08X-%04X-%04X-... (data1 shown)\n", prof.Clsid.Data1, prof.Clsid.Data2, prof.Clsid.Data3)
		return
	}
	// Fallback to classic GetKeyboardLayoutName
	klid, err := getActiveKLIDClassic()
	if err == nil && klid != "" {
		// lookup display name from registry
		regs, _ := readRegisteredKeyboardLayouts()
		name := ""
		if info, ok := regs[klid]; ok {
			name = info.Name
		}
		fmt.Printf(" - KLID: %s\n", klid)
		if name != "" {
			fmt.Printf(" - Display name: %s\n", name)
		}
	} else {
		fmt.Println(" - Could not determine current input method (no TSF profile, GetKeyboardLayoutName failed)")
	}
}

func main() {
	if len(os.Args) == 1 {
		// no args: show current input method (no help)
		printStatus()
		return
	}
	if len(os.Args) >= 2 {
		arg := os.Args[1]
		switch arg {
		case "--help", "-h", "help":
			printHelp()
			return
		case "--list", "-l":
			fmt.Println("Listing candidates (registry keyboard layouts + detected TSF profiles)...")
			cands, _ := listCandidates()
			for _, c := range cands {
				fmt.Printf("%-5s %-10s (lang 0x%04X) %-20s %s\n", c.Source, c.KLID, c.Langid, c.TypeDesc, c.Display)
			}
			return
		}
		// try to interpret argument: KLID, HKL, LANGID, or NAME_SUBSTR
		// heuristics:
		a := arg
		// If it's exactly 8 hex digits -> treat as KLID
		if ok := isEightHex(a); ok {
			if err := switchByKLID(a); err != nil {
				fmt.Fprintf(os.Stderr, "switch failed: %v\n", err)
				os.Exit(2)
			}
			return
		}
		// If looks like 0xHEX or contains hex letters and length between 3..10 -> HKL
		if strings.HasPrefix(strings.ToLower(a), "0x") || stringsContainsHexLetters(a) {
			if err := switchByHKLStr(a); err != nil {
				fmt.Fprintf(os.Stderr, "switch failed: %v\n", err)
				os.Exit(2)
			}
			return
		}
		// if numeric small -> LANGID
		if _, err := strconv.ParseUint(a, 10, 64); err == nil {
			if err := switchByLangIDStr(a); err == nil {
				return
			}
			// else fallthrough to name substring
		}
		// fallback: treat as name substring
		if err := switchByNameSubstr(a); err != nil {
			fmt.Fprintf(os.Stderr, "switch failed: %v\n", err)
			os.Exit(2)
		}
	}
}

// helpers for arg detection
func isEightHex(s string) bool {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		s = s[2:]
	}
	if len(s) != 8 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
func stringsContainsHexLetters(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f') {
			return true
		}
	}
	return false
}
