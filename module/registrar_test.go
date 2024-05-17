package module

import (
	"fmt"
	"net"
	"testing"
)

func TestRegNew(t *testing.T) {
	registrar := NewRegistrar()
	if registrar == nil {
		t.Fatal("Could not create registrar")
	}
}

func TestRegRegistrar(t *testing.T) {
	mt := TYPE_DOWNLOADER
	ml := legalTypeLetterMap[mt]
	sn := DefaultSNGen.Get()
	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))
	registrar := NewRegistrar()
	ok, err := registrar.Register(nil)
	if err == nil {
		t.Fatal("No error when register module instance with nil module")
	}
	if ok {
		t.Fatal("It still can register nil module instance")
	}
	var m Module
	for t, f := range fakeModuleFuncMap {
		if t != mt {
			m = f(mid)
			break
		}
	}
	ok, err = registrar.Register(m)
	if err == nil {
		t.Fatalf("No error when register unmatched module instance, (type: %T)", m)
	}
	if ok {
		t.Fatalf("It still can register unmatched module instance, (type: %T)", m)
	}
	var midsAll []MID
	for _, mt := range legalTypes {
		var midsByType []MID
		for mip := range legalIPMap {
			ml = legalTypeLetterMap[mt]
			sn = DefaultSNGen.Get()
			addr, _ = NewAddr("http", mip, 8080)
			mid = MID(fmt.Sprintf(midTemplate, ml, sn, addr))
			midsByType = append(midsByType, mid)
			midsAll = append(midsAll, mid)
			m = fakeModuleFuncMap[mt](mid)
			ok, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s (MID: %s)", err, mid)
			}
			if !ok {
				t.Fatalf("Could not register module instance with MID %q", mid)
			}
			ok, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s (MID: %s)", err, mid)
			}
			if ok {
				t.Fatalf("It still can repeatedly register module instance with same MID %q", mid)
			}
			sn = DefaultSNGen.Get()
			mid = MID(fmt.Sprintf(midTemplate, "M", sn, addr))
			m = fakeModuleFuncMap[mt](mid)
			ok, err = registrar.Register(m)
			if err == nil {
				t.Fatalf("No error when registering module instance with illegal MID %q", mid)
			}
			if ok {
				t.Fatalf("It can still register module install with illegal MID %q", mid)
			}
		}
		modules, err := registrar.GetAllByType(mt)
		if err != nil {
			t.Fatalf("An error occurs when getting all module instances: %s, (type: %s)", err, mt)
		}
		for _, mid := range midsByType {
			if _, ok := modules[mid]; !ok {
				t.Fatalf("Not found the module instance (MID: %s, type: %s)", mid, mt)
			}
		}
	}
	modules := registrar.GetAll()
	for _, mid := range midsAll {
		if _, ok := modules[mid]; !ok {
			t.Fatalf("Not found the module instance (MID: %s)", mid)
		}
	}
	for _, mt := range illegalTypes {
		sn := DefaultSNGen.Get()
		addr, _ := NewAddr("http", "127.0.0.1", 8080)
		ml := legalTypeLetterMap[mt]
		mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))
		m := NewFakeDownloader(mid, nil)
		ok, err := registrar.Register(m)
		if err == nil {
			t.Fatalf("No error when register module instance with illegal type %q", mt)
		}
		if ok {
			t.Fatalf("It still can register module instance with illegal type %q", mt)
		}
	}
}

func TestModuleUnregister(t *testing.T) {
	registrar := NewRegistrar()
	var mids []MID
	for _, mt := range legalTypes {
		for mip := range legalIPMap {
			sn := DefaultSNGen.Get()
			addr, _ := NewAddr("http", mip, 8080)
			mid, err := GenMID(mt, sn, addr)
			if err != nil {
				t.Fatalf("An error occurs when generating module ID: %s (type: %s, sn: %d, addr: %s)",
					err, mt, sn, addr)
			}
			m := fakeModuleFuncMap[mt](mid)
			_, err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s, (type: %s, sn: %d, addr: %s)",
					err, mt, sn, addr)
			}
			mids = append(mids, mid)
		}
	}
	for _, mid := range mids {
		ok, err := registrar.Unregister(mid)
		if err != nil {
			t.Fatalf("An error occurs when unregistering module instance %s", mid)
		}
		if !ok {
			t.Fatalf("Could not unregister module instance (MID: %s)", mid)
		}
	}
	for _, mid := range mids {
		ok, err := registrar.Unregister(mid)
		if err != nil {
			t.Fatalf("An error occurs when unregistering module instance %s", mid)
		}
		if ok {
			t.Fatalf("It can still unregister nonexist module instance (MID: %s)", mid)
		}
	}
	for _, illegalMID := range illegalMIDs {
		ok, err := registrar.Unregister(illegalMID)
		if err == nil {
			t.Fatalf("No error occurs when unregistering module instance with illegal MID %q", illegalMID)
		}
		if ok {
			t.Fatalf("It still unregister module instance with illegal MID: %q", illegalMID)
		}
	}
}

func TestModuleGet(t *testing.T) {
	registrar := NewRegistrar()
	mt := illegalTypes[0]
	m1, err := registrar.Get(mt)
	if err == nil {
		t.Fatalf("No error when get module instance with illegal type %q", mt)
	}
	if m1 != nil {
		t.Fatalf("It still can get module instance with illegal type %q", mt)
	}
	mt = TYPE_DOWNLOADER
	m1, err = registrar.Get(mt)
	if err == nil {
		t.Fatalf("No error when get nonexist module instance")
	}
	if m1 != nil {
		t.Fatalf("It still can get nonexist module instance")
	}
	//addr, _ := NewAddr("http", "127.0.0.1", 8080)
	//	mid := MID(fmt.Sprintf(midTemplate, legalTypeLetterMap[mt], DefaultSNGen.Get(), addr))
	m := defaultFakeModuleMap[mt]
	_, err = registrar.Register(m)
	if err != nil {
		t.Fatalf("An error occurs when registering module instance: %s (mid: %s)", err, m.ID())
	}
	m1, err = registrar.Get(mt)
	if err != nil {
		t.Fatalf("An error occurs when getting module instance: %s, (MID: %s)", err, m.ID())
	}
	if m1 == nil {
		t.Fatalf("Could not get module instance with MID %q", m.ID())
	}
	if m1.ID() != m.ID() {
		t.Fatalf("Inconsistent MID, expected: %s, actual: %s", m.ID(), m1.ID())
	}
}

func TestModuleAllInParallel(t *testing.T) {
	baseSize := 1000
	basePort := 8000
	legalTypesLen := len(legalTypes)
	sLen := baseSize * legalTypesLen
	types := make([]Type, sLen)
	sns := make([]uint64, sLen)
	addrs := make([]net.Addr, sLen)
	for i := 0; i < sLen; i++ {
		types[i] = legalTypes[i%legalTypesLen]
		port := uint64(basePort + i%legalTypesLen)
		addrs[i], _ = NewAddr("http", "127.0.0.1", port)
		sns[i] = DefaultSNGen.Get()
	}
	registrar := NewRegistrar()
	t.Run("All in Parallel", func(t *testing.T) {
		t.Run("Register", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := DefaultSNGen.Get()
				mid, err := GenMID(mt, sn, addr)
				if err != nil {
					t.Fatalf("An error occurs when generating module ID: %s, (type: %s, sn: %d, addr: %s)",
						err, mt, sn, addr)
				}
				m := fakeModuleFuncMap[mt](mid)
				_, err = registrar.Register(m)
				if err != nil {
					t.Fatalf("An error occurs when registering module instance: %s, (type: %s, sn: %d, addr: %s)",
						err, mt, sn, addr)
				}
			}
		})
		t.Run("Unregister", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := sns[i]
				mid, _ := GenMID(mt, sn, addr)
				_, err := registrar.Unregister(mid)
				if err != nil {
					t.Fatalf("An error occurs when unregistering module instance: %s, (MID: %s)", err, mid)
				}
			}
		})
		t.Run("Get", func(t *testing.T) {
			t.Parallel()
			for _, mt := range types {
				m, err := registrar.Get(mt)
				if err != nil && err != ErrNotFoundModuleInstance {
					t.Fatalf("An error occurs when gettting module instance: %s (mtype: %s, MID: %s)", err, mt, m.ID())
				}
			}
		})
	})
}
