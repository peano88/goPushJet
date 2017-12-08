package goPushJet

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type testParams struct {
	ID        int
	Service   string
	Operation string
	Secret    string
	Name      string
	Icon      string
	Errorf    bool                      // An error is expected
	Servicef  bool                      // A real Service is expected
	Special   func(Service, *testing.T) // Special Check, which can be implemented
	// or not
}

func createTestLandscape(existing Service) []testParams {

	if existing.Secret == "" || existing.Public == "" {
		return []testParams{}
	}
	fmt.Println("Secret:" + existing.Secret)
	fmt.Println("Public:" + existing.Public)
	return []testParams{
		//GetInfo Success
		testParams{
			ID:        1,
			Operation: "Read",
			Service:   existing.Public,
			Secret:    existing.Secret,
			Errorf:    false,
			Servicef:  true,
			Special: func(s Service, t *testing.T) {
				if s.Name != existing.Name {
					t.Errorf("Read Information went wrong")
				}

			},
		},
		//GetInfo Error
		testParams{
			ID:        2,
			Operation: "Read",
			Service:   "NOT_EXISTING",
			Errorf:    true,
			Servicef:  false,
		},
		//Update Success
		testParams{
			ID:        3,
			Operation: "Update",
			Service:   existing.Public,
			Secret:    existing.Secret,
			Name:      "Second",
			Errorf:    false,
			Servicef:  false,
		},
		// CHeck if the Update was ok
		testParams{
			ID:        30,
			Operation: "Read",
			Service:   existing.Public,
			Secret:    existing.Secret,
			Errorf:    false,
			Servicef:  true,
			Special: func(s Service, t *testing.T) {
				if s.Name != "Second" {
					t.Errorf("Update went wrong")
				}

			},
		},
		//Update Error
		testParams{
			ID:        4,
			Operation: "Update",
			Secret:    "NOT_EXISTING",
			Errorf:    true,
			Servicef:  false,
		},
		//Delete Success
		testParams{
			ID:        5,
			Operation: "Delete",
			Secret:    existing.Secret,
			Errorf:    false,
			Servicef:  false,
		},
		//Delete Error}
		testParams{
			ID:        6,
			Operation: "Delete",
			Secret:    "NOT_EXISTING",
			Errorf:    true,
			Servicef:  false,
		},
	}
}

func callFunction(tp testParams) (Service, error) {
	switch tp.Operation {
	case "Create":
		return CreateService(tp.Name, tp.Icon)
	case "Update":
		return Service{}, UpdateService(tp.Secret, tp.Name, tp.Icon)
	case "Delete":
		return Service{}, DeleteService(tp.Secret)
	case "Read":
		return GetServiceInfo(tp.Service, tp.Secret)
	}

	return Service{}, nil

}

func checkError(err error, tp testParams, t *testing.T) {
	if tp.Errorf == true && err == nil {
		t.Errorf(" For test %v an error is expected", tp.ID)
	}

	if tp.Errorf == false && err != nil {
		t.Errorf("For test %v an error is not expected: "+err.Error(), tp.ID)
	}
}

func checkService(s Service, tp testParams, t *testing.T) {
	if tp.Servicef == true && s.IsEmpty() {
		t.Errorf(" For test %v a service is expected", tp.ID)
	}

	if tp.Servicef == false && !s.IsEmpty() {
		t.Errorf("For test %v a service is not expected", tp.ID)
	}
}

func checkSpecial(s Service, tp testParams, t *testing.T) {
	if tp.Special != nil {
		tp.Special(s, t)
	}
}

func createTestService(s *Service, f *os.File) (e error) {
	*s, e = CreateService("first", "") //no icon needed
	if e != nil {
		return e
	}

	e = json.NewEncoder(f).Encode(*s)
	if e != nil {
		return e
	}

	e = bufio.NewWriter(f).Flush()
	if e != nil {
		return e
	}
	return nil
}

func TestService(t *testing.T) {
	//open the file with JSON data over a created service
	ts := &Service{}
	f, err := os.Open("serialized_valid_service.txt")
	defer f.Close()
	if err != nil {
		f, err = os.Create("serialized_valid_service.txt")
		if err != nil {
			t.Fatalf("No Test Service Serialization available")
		}

		err = createTestService(ts, f)
		if err != nil || ts.IsEmpty() {
			t.Fatalf(err.Error())
		}

	} else {

		e := json.NewDecoder(f).Decode(ts)

		if e != nil {
			t.Fatalf("marshal.error:" + e.Error())
			//create a First Service to have a valid service/secret reference
		}
	}
	tl := createTestLandscape(*ts)

	if len(tl) == 0 {
		t.Fatalf("No test landscape created")
	}

	for _, tp := range tl {
		s, err := callFunction(tp)
		checkService(s, tp, t)
		checkError(err, tp, t)
	}

}
