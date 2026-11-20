package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeDesktopCodeStore struct {
	codes map[string]DesktopConnectCode
	used  map[string]bool
}

func newFakeDesktopCodeStore() *fakeDesktopCodeStore {
	return &fakeDesktopCodeStore{codes: map[string]DesktopConnectCode{}, used: map[string]bool{}}
}

func (s *fakeDesktopCodeStore) Save(_ context.Context, codeHash string, code DesktopConnectCode, _ time.Duration) error {
	s.codes[codeHash] = code
	return nil
}

func (s *fakeDesktopCodeStore) Consume(_ context.Context, codeHash string) (*DesktopConnectCode, error) {
	if code, ok := s.codes[codeHash]; ok {
		delete(s.codes, codeHash)
		s.used[codeHash] = true
		return &code, nil
	}
	if s.used[codeHash] {
		return nil, ErrDesktopCodeAlreadyUsed
	}
	return nil, ErrDesktopCodeExpired
}

type fakeAppDeviceTokenRepo struct {
	owners  map[string]int64
	devices map[string]string
	saved   AppDevice
}

func newFakeAppDeviceTokenRepo() *fakeAppDeviceTokenRepo {
	return &fakeAppDeviceTokenRepo{owners: map[string]int64{}, devices: map[string]string{}}
}

func (r *fakeAppDeviceTokenRepo) Replace(_ context.Context, userID int64, device AppDevice, tokenHash string) error {
	if old, ok := r.devices[device.DeviceID]; ok {
		delete(r.owners, old)
	}
	r.devices[device.DeviceID] = tokenHash
	r.owners[tokenHash] = userID
	r.saved = device
	return nil
}

func (r *fakeAppDeviceTokenRepo) ActiveUserID(_ context.Context, tokenHash string) (int64, bool, error) {
	owner, ok := r.owners[tokenHash]
	return owner, ok, nil
}

func TestCanonicalDesktopOrigin(t *testing.T) {
	valid := map[string]string{
		"https://yeschoy.com":  "https://yeschoy.com",
		"https://YesChoy.com/": "https://yeschoy.com",
	}
	for in, want := range valid {
		got, err := CanonicalDesktopOrigin(in)
		if err != nil || got != want {
			t.Errorf("CanonicalDesktopOrigin(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	invalid := []string{
		"", "http://yeschoy.com", "https://user:pw@yeschoy.com", "https://yeschoy.com:443", "https://yeschoy.com:8443",
		"https://yeschoy.com/path", "https://yeschoy.com?x=1", "https://yeschoy.com#frag", "https://127.0.0.1", "https://[::1]", "yeschoy.com",
	}
	for _, in := range invalid {
		if _, err := CanonicalDesktopOrigin(in); !errors.Is(err, ErrDesktopOriginInvalid) {
			t.Errorf("CanonicalDesktopOrigin(%q) error = %v; want ErrDesktopOriginInvalid", in, err)
		}
	}
}

type fakeDesktopUserRepo struct {
	UserRepository
	users map[int64]*User
}

func (r *fakeDesktopUserRepo) GetByID(_ context.Context, id int64) (*User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func newFakeDesktopUserRepo() *fakeDesktopUserRepo {
	return &fakeDesktopUserRepo{users: map[int64]*User{
		7: {ID: 7, Email: "u7@example.com", PasswordHash: "hash-before", Status: StatusActive},
	}}
}

func TestDesktopConnectIssueCodeRejectsInvalidOrigin(t *testing.T) {
	svc := NewDesktopConnectService(newFakeDesktopCodeStore(), newFakeAppDeviceTokenRepo(), nil, nil)
	if _, err := svc.IssueCode(context.Background(), 1, "http://yeschoy.com"); !errors.Is(err, ErrDesktopOriginInvalid) {
		t.Fatalf("IssueCode error = %v; want ErrDesktopOriginInvalid", err)
	}
}

// A code asked for before a password change must not sign anyone in after it.
func TestDesktopConnectExchangeRejectsCodeIssuedBeforePasswordChange(t *testing.T) {
	users := newFakeDesktopUserRepo()
	svc := NewDesktopConnectService(newFakeDesktopCodeStore(), newFakeAppDeviceTokenRepo(), users, nil)
	ctx := context.Background()
	issued, err := svc.IssueCode(ctx, 7, "https://yeschoy.com")
	if err != nil {
		t.Fatalf("IssueCode: %v", err)
	}
	users.users[7].PasswordHash = "hash-after"
	if _, err := svc.Exchange(ctx, issued.Code, "yeschoy.com"); !errors.Is(err, ErrDesktopCodeExpired) {
		t.Fatalf("Exchange after a password change error = %v; want ErrDesktopCodeExpired", err)
	}
}

func TestDesktopConnectExchangeRejectsWrongHostAndBurnsTheCode(t *testing.T) {
	svc := NewDesktopConnectService(newFakeDesktopCodeStore(), newFakeAppDeviceTokenRepo(), newFakeDesktopUserRepo(), nil)
	ctx := context.Background()
	issued, err := svc.IssueCode(ctx, 7, "https://yeschoy.com")
	if err != nil {
		t.Fatalf("IssueCode: %v", err)
	}
	if len(issued.Code) < 32 || issued.Origin != "https://yeschoy.com" {
		t.Fatalf("issued = %+v", issued)
	}
	if _, err := svc.Exchange(ctx, issued.Code, "evil.example:443"); !errors.Is(err, ErrDesktopOriginMismatch) {
		t.Fatalf("Exchange on another host error = %v; want ErrDesktopOriginMismatch", err)
	}
	if _, err := svc.Exchange(ctx, issued.Code, "yeschoy.com"); !errors.Is(err, ErrDesktopCodeAlreadyUsed) {
		t.Fatalf("second Exchange error = %v; want ErrDesktopCodeAlreadyUsed", err)
	}
	if _, err := svc.Exchange(ctx, "never-issued-code-never-issued-code-0000", "yeschoy.com"); !errors.Is(err, ErrDesktopCodeExpired) {
		t.Fatalf("unknown code error = %v; want ErrDesktopCodeExpired", err)
	}
}

func TestDesktopAppTokenBelongsToOneUserAndOneDevice(t *testing.T) {
	repo := newFakeAppDeviceTokenRepo()
	svc := NewDesktopConnectService(newFakeDesktopCodeStore(), repo, nil, nil)
	ctx := context.Background()

	if _, err := svc.IssueAppToken(ctx, 7, AppDevice{DeviceID: "  "}); !errors.Is(err, ErrDesktopDeviceInvalid) {
		t.Fatalf("blank device error = %v; want ErrDesktopDeviceInvalid", err)
	}

	first, err := svc.IssueAppToken(ctx, 7, AppDevice{DeviceID: "device-a", Platform: "amiga"})
	if err != nil {
		t.Fatalf("IssueAppToken: %v", err)
	}
	if repo.saved.Platform != "unknown" {
		t.Errorf("platform = %q; want unknown", repo.saved.Platform)
	}
	if !svc.IsAppRequest(ctx, first, 7) {
		t.Error("the owner's token should mark the request as coming from the app")
	}
	if svc.IsAppRequest(ctx, first, 8) {
		t.Error("another user's token must not count")
	}
	if svc.IsAppRequest(ctx, "", 7) || svc.IsAppRequest(ctx, "not-an-app-token", 7) {
		t.Error("a missing or malformed token must not count")
	}

	second, err := svc.IssueAppToken(ctx, 7, AppDevice{DeviceID: "device-a", Platform: "macos"})
	if err != nil {
		t.Fatalf("IssueAppToken again: %v", err)
	}
	if svc.IsAppRequest(ctx, first, 7) || !svc.IsAppRequest(ctx, second, 7) {
		t.Error("connecting the same device again must replace its previous token")
	}
}
