package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type LoginResponse struct {
	RequireOTP   bool   `json:"otpRequired"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userId"`
}

func TestMain(m *testing.M) {
	os.Setenv("POSTGRES_URL", "postgres://postgres:root@localhost/flexspace_test?sslmode=disable")
	os.Setenv("MOCK_SENDMAIL", "1")
	os.Setenv("FASTSPRING_VALIDATE", "0")
	GetConfig().ReadConfig()
	db := GetDatabase()
	dropTestDB()
	a := GetApp()
	a.InitializeDatabases()
	a.InitializeRouter()
	code := m.Run()
	db.Close()
	os.Exit(code)
}

func getTestJWT(userID string) string {
	claims := &Claims{
		Email:  userID,
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(60 * 24 * 14 * time.Minute).Unix(),
		},
	}
	router := &AuthRouter{}
	jwt := router.createJWT(claims)
	return jwt
}

func newHTTPRequest(method, url, userID string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if userID != "" {
		req.Header.Set("Authorization", "Bearer "+getTestJWT(userID))
	}
	return req
}

func createTestUser(orgDomain string) *User {
	return createTestUserParams(orgDomain)
}

func createTestUserParams(orgDomain string) *User {
	org := createTestOrg(orgDomain)
	user := &User{
		Email:          uuid.New().String() + "@" + orgDomain,
		OrganizationID: org.ID,
		OrgAdmin:       false,
		SuperAdmin:     false,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func createTestUserSuperAdmin() *User {
	org := createTestOrg("test.com")
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: org.ID,
		OrgAdmin:       false,
		SuperAdmin:     true,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func createTestOrg(orgDomain string) *Organization {
	org := &Organization{
		Name:             "Test Org",
		ContactEmail:     "foo@seatsurfing.de",
		ContactFirstname: "Foo",
		ContactLastname:  "Bar",
		Country:          "DE",
		Language:         "de",
	}
	if err := GetOrganizationRepository().Create(org); err != nil {
		panic(err)
	}
	if err := GetOrganizationRepository().AddDomain(org, orgDomain, true); err != nil {
		panic(err)
	}
	return org
}

func createTestUserInOrg(org *Organization) *User {
	user := &User{
		Email:          uuid.New().String() + "@test.com",
		OrganizationID: org.ID,
		OrgAdmin:       false,
		SuperAdmin:     false,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func createTestUserOrgAdminDomain(org *Organization, domain string) *User {
	user := &User{
		Email:          uuid.New().String() + "@" + domain,
		OrganizationID: org.ID,
		OrgAdmin:       true,
		SuperAdmin:     false,
	}
	if err := GetUserRepository().Create(user); err != nil {
		panic(err)
	}
	return user
}

func createTestUserOrgAdmin(org *Organization) *User {
	return createTestUserOrgAdminDomain(org, "test.com")
}

func loginTestUserParams(userID string) *LoginResponse {
	// TODO
	res := &LoginResponse{
		AccessToken:  "abc",
		RefreshToken: "def",
		RequireOTP:   false,
		UserID:       userID,
	}
	return res
}

func loginTestUser(userID string) *LoginResponse {
	return loginTestUserParams(userID)
}

func createLoginTestUser() *LoginResponse {
	user := createTestUser("test.com")
	return loginTestUser(user.ID)
}

func createLoginTestUserParams() *LoginResponse {
	user := createTestUserParams("test.com")
	return loginTestUserParams(user.ID)
}

func dropTestDB() {
	tables := []string{"auth_providers", "auth_states", "bookings", "spaces", "locations", "organizations_domains", "organizations", "users", "signups", "settings", "subscription_events"}
	for _, s := range tables {
		GetDatabase().DB().Exec("DROP TABLE IF EXISTS " + s)
	}
}

func clearTestDB() {
	tables := []string{"auth_providers", "auth_states", "bookings", "spaces", "locations", "organizations_domains", "organizations", "users", "signups", "settings", "subscription_events"}
	for _, s := range tables {
		GetDatabase().DB().Exec("TRUNCATE " + s)
	}
}

func executeTestRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	GetApp().Router.ServeHTTP(rr, req)
	return rr
}

func checkTestResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected HTTP Status %d, but got %d", expected, actual)
	}
}

func checkTestString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("Expected '%s', but got '%s'", expected, actual)
	}
}

func checkTestBool(t *testing.T, expected, actual bool) {
	if expected != actual {
		t.Fatalf("Expected '%t', but got '%t'", expected, actual)
	}
}

func checkTestUint(t *testing.T, expected, actual uint) {
	if expected != actual {
		t.Fatalf("Expected '%d', but got '%d'", expected, actual)
	}
}

func checkTestInt(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Fatalf("Expected '%d', but got '%d'", expected, actual)
	}
}

func checkStringNotEmpty(t *testing.T, s string) {
	if strings.TrimSpace(s) == "" {
		t.Fatalf("Expected non-empty string")
	}
}