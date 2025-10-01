package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

func givenThereIsADefaultClientBoootstrapped() error {
	// assumed to be true
	return nil
}

func givenThereIsAnAdminAccountBootstrapped() error {
	// assumed to be true
	return nil
}

func givenBootstrapAccountCredentialsAreValid() error {
	// passed in the default context
	return nil
}

type tokenResponseKey struct{}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	StatusCode   int
}

func (tr *tokenResponse) authHeaderValue() string {
	return fmt.Sprintf("%s %s", tr.TokenType, tr.AccessToken)
}

func authenthicateWithCredentialsViaClientCredentialsFlow(ctx context.Context, userCreds userCredentials, clientCreds clientCredentials) (context.Context, error) {
	tokenEndpoint, err := makeRequestURL(ctx, "/oauth/v2/token")
	if err != nil {
		return ctx, err
	}
	client := http.Client{}
	formBody := strings.NewReader(url.Values{
		"username":      []string{userCreds.username},
		"password":      []string{userCreds.password},
		"client_id":     []string{clientCreds.id},
		"client_secret": []string{clientCreds.secret},
		"grant_type":    []string{"password"},
	}.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, formBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return ctx, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}
	logger.DebugContext(ctx, "Received response", "statusCode", resp.StatusCode, "body", string(body), "headers", resp.Header)
	response := tokenResponse{
		StatusCode: resp.StatusCode,
	}
	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return ctx, err
		}
	}
	return context.WithValue(ctx, tokenResponseKey{}, response), nil
}

func givenIAmAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}
	return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, bootstrapCreds, bootstrapClient)
}

func thenIAmSuccessfullyAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("unable to obtain token response")
	}
	return ctx, nil
}

func thenTheAccountCreationShouldSucceed(ctx context.Context) (context.Context, error) {
	createdAccount, ok := ctx.Value(createdAccountKey{}).(createdAccountResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing created account")
	}

	if createdAccount.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("account creation failed with status %d", createdAccount.StatusCode)
	}

	if createdAccount.Username == "" {
		return ctx, fmt.Errorf("created account has empty username")
	}

	if createdAccount.Password == "" {
		return ctx, fmt.Errorf("created account has empty password")
	}

	return ctx, nil
}

func thenIAmAbleToLoginAsThatAccount(ctx context.Context) (context.Context, error) {
	createdAccount, ok := ctx.Value(createdAccountKey{}).(createdAccountResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing created account")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}

	newUserCreds := userCredentials{
		username: createdAccount.Username,
		password: createdAccount.Password,
	}

	ctx, err := authenthicateWithCredentialsViaClientCredentialsFlow(ctx, newUserCreds, bootstrapClient)
	if err != nil {
		return ctx, fmt.Errorf("failed to authenticate with newly created account: %w", err)
	}

	// Verify we got a valid token
	tokenResp, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("failed to obtain token response after authentication")
	}

	if tokenResp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("authentication failed with status %d", tokenResp.StatusCode)
	}

	return ctx, nil
}

type createdAccountKey struct{}

type createdAccountResponse struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	StatusCode int
}

func whenICreateANewAccount(ctx context.Context, accountType string) (context.Context, error) {
	token, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing token response")
	}

	createUserEndpoint, err := makeRequestURL(ctx, "/api/users")
	if err != nil {
		return ctx, err
	}

	username := fmt.Sprintf("test-%s-%d", accountType, ctx.Value("timestamp"))
	email := fmt.Sprintf("%s@wallabago.local", username)

	formValues := url.Values{
		"username": []string{username},
		"email":    []string{email},
	}

	// Set is_admin based on account type
	if accountType == "admin" {
		formValues.Set("is_admin", "true")
	}

	formBody := strings.NewReader(formValues.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createUserEndpoint, formBody)
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", token.authHeaderValue())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	logger.DebugContext(ctx, "Received create account response", "statusCode", resp.StatusCode, "body", string(body))

	response := createdAccountResponse{
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return ctx, err
		}
	}

	return context.WithValue(ctx, createdAccountKey{}, response), nil
}

func makeRequestURL(ctx context.Context, path string) (string, error) {
	addr, ok := ctx.Value(serverAddrKey{}).(string)
	if !ok {
		return "", fmt.Errorf("failed to extract server address from context")
	}
	return fmt.Sprintf("http://%s%s", addr, path), nil
}

func whenIUseBootstrapCredentialsToAuthenticate(ctx context.Context) (context.Context, error) {
	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}
	return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, bootstrapCreds, bootstrapClient)
}

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	rand.Seed(time.Now().UnixNano())
}

type myUserAccountKey struct{}
type anotherUserAccountKey struct{}

func generateRandomSuffix() string {
	return fmt.Sprintf("%d", rand.Int63n(1000000))
}

func createUserAccountViaAdmin(ctx context.Context, accountName string) (userCredentials, error) {
	// Authenticate as bootstrap admin
	bootstrapCreds, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return userCredentials{}, fmt.Errorf("failed to extract bootstrap credentials")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return userCredentials{}, fmt.Errorf("failed to extract bootstrap client")
	}

	adminCtx, err := authenthicateWithCredentialsViaClientCredentialsFlow(ctx, bootstrapCreds, bootstrapClient)
	if err != nil {
		return userCredentials{}, err
	}

	adminToken, ok := adminCtx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return userCredentials{}, fmt.Errorf("failed to get admin token")
	}

	// Create user account via admin
	createUserEndpoint, err := makeRequestURL(ctx, "/api/users")
	if err != nil {
		return userCredentials{}, err
	}

	email := fmt.Sprintf("%s@wallabago.local", accountName)

	formBody := strings.NewReader(url.Values{
		"username": []string{accountName},
		"email":    []string{email},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, createUserEndpoint, formBody)
	if err != nil {
		return userCredentials{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", adminToken.authHeaderValue())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return userCredentials{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return userCredentials{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return userCredentials{}, fmt.Errorf("failed to create user account: status %d, body: %s", resp.StatusCode, string(body))
	}

	var createdUser createdAccountResponse
	err = json.Unmarshal(body, &createdUser)
	if err != nil {
		return userCredentials{}, fmt.Errorf("failed to parse created user: %v, body: %s", err, string(body))
	}

	return userCredentials{
		username: createdUser.Username,
		password: createdUser.Password,
	}, nil
}

func givenThereExistsMyUserAccount(ctx context.Context) (context.Context, error) {
	accountName := fmt.Sprintf("my-account-%s", generateRandomSuffix())
	creds, err := createUserAccountViaAdmin(ctx, accountName)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, myUserAccountKey{}, creds), nil
}

func givenThereExistsAnotherUserAccount(ctx context.Context) (context.Context, error) {
	accountName := fmt.Sprintf("another-account-%s", generateRandomSuffix())
	creds, err := createUserAccountViaAdmin(ctx, accountName)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, anotherUserAccountKey{}, creds), nil
}

func givenThereExistsAClient(ctx context.Context) (context.Context, error) {
	// well use the bootstrapped client for now
	return ctx, nil
}

type (
	clientCredentialsKey struct{}
	userCredentialsKey   struct{}
)

func givenClientCredentialsValidity(ctx context.Context, validity string) (context.Context, error) {
	var client clientCredentials

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}

	switch validity {
	case "valid":
		client = bootstrapClient
	case "invalid":
		client = clientCredentials{
			id:     bootstrapClient.id,
			secret: bootstrapClient.secret + "-some-junk",
		}
	default:
		return ctx, fmt.Errorf("bad value for validity: %s", validity)
	}

	return context.WithValue(ctx, clientCredentialsKey{}, client), nil
}

func givenUserCredentialsValidity(ctx context.Context, validity string) (context.Context, error) {
	var user userCredentials

	bootstrapUser, ok := ctx.Value(bootstrapCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap user")
	}

	switch validity {
	case "valid":
		user = bootstrapUser
	case "invalid":
		user = userCredentials{
			username: bootstrapUser.username,
			password: bootstrapUser.password + "-some-junk",
		}
	default:
		return ctx, fmt.Errorf("bad value for validity: %s", validity)
	}

	return context.WithValue(ctx, userCredentialsKey{}, user), nil
}

func whenClientUsesCredentialsToAuthenticate(ctx context.Context) (context.Context, error) {
	user, ok := ctx.Value(userCredentialsKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract user credentials")
	}

	client, ok := ctx.Value(clientCredentialsKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract client credentials")
	}
	return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, user, client)
}

func thenTheClientAuthOutcomeShouldBe(ctx context.Context, outcome string) (context.Context, error) {
	actualOutcome, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("failed to extract auth outcome from context")
	}

	switch outcome {
	case "authenticated":
		if actualOutcome.StatusCode != http.StatusOK {
			return ctx, fmt.Errorf("auth outcome should succeed, instead got %d status code", actualOutcome.StatusCode)
		}
	case "rejected":
		if actualOutcome.StatusCode != http.StatusUnauthorized {
			return ctx, fmt.Errorf("auth outcome should fail, instead got %d status code", actualOutcome.StatusCode)
		}
	default:
		return ctx, fmt.Errorf("invalid value for outcome: %s", outcome)
	}
	return ctx, nil
}

type entryURLKey struct{}

func givenEntryURLPointsToHTMLPage(ctx context.Context, validity string) (context.Context, error) {
	var pageURL string
	switch validity {
	case "valid":
		pageURL = "https://en.wikipedia.org/wiki/Behavior-driven_development"
	case "invalid":
		pageURL = "https://this.domain.should.not.exist/and/this/path/also"
	default:
		return ctx, fmt.Errorf("bad validity: %s", validity)
	}
	return context.WithValue(ctx, entryURLKey{}, pageURL), nil
}

func givenIAmAuthenticatedAsUser(ctx context.Context) (context.Context, error) {
	// Use "my" user account
	myUserCreds, ok := ctx.Value(myUserAccountKey{}).(userCredentials)
	if !ok {
		return ctx, fmt.Errorf("my user account not found in context")
	}

	bootstrapClient, ok := ctx.Value(bootstrapClientKey{}).(clientCredentials)
	if !ok {
		return ctx, fmt.Errorf("failed to extract bootstrap client")
	}
	return authenthicateWithCredentialsViaClientCredentialsFlow(ctx, myUserCreds, bootstrapClient)
}

type addEntryResponseKey struct{}

type addEntryResult struct {
	StatusCode int
	Body       []byte
}

func whenITryToAddAnEntry(ctx context.Context) (context.Context, error) {
	pageURL, ok := ctx.Value(entryURLKey{}).(string)
	if !ok {
		return ctx, fmt.Errorf("context is missing entry url")
	}
	token, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing token response")
	}
	entryEndpoint, err := makeRequestURL(ctx, "/api/entries")
	if err != nil {
		return ctx, err
	}
	formBody := strings.NewReader(url.Values{
		"url": []string{pageURL},
	}.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, entryEndpoint, formBody)
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", token.authHeaderValue())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	result := addEntryResult{
		StatusCode: resp.StatusCode,
		Body:       body,
	}
	return context.WithValue(ctx, addEntryResponseKey{}, result), nil
}

func thenEntryAdditionShouldHaveStatus(ctx context.Context, success string) (context.Context, error) {
	response, ok := ctx.Value(addEntryResponseKey{}).(addEntryResult)
	if !ok {
		return ctx, fmt.Errorf("context is missing add entry result")
	}
	switch success {
	case "should":
		if response.StatusCode != http.StatusOK {
			return ctx, fmt.Errorf("bad status code, expected 200 but got %d; body: %s", response.StatusCode, string(response.Body))
		}
	case "should not":
		//nolint:usestdlibvars // false-positive for 100 -> http.StatusContinue
		if (response.StatusCode / 100) != 4 {
			return ctx, fmt.Errorf("bad status code, expected 4xx but got %d; body: %s", response.StatusCode, string(response.Body))
		}
	}
	return ctx, nil
}

func thenTheEntryShouldHaveExistence(ctx context.Context, existence string) (context.Context, error) {
	pageURL, ok := ctx.Value(entryURLKey{}).(string)
	if !ok {
		return ctx, fmt.Errorf("context is missing entry url")
	}
	tokenResponse, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing token response")
	}

	existenceEndpoint, err := makeRequestURL(ctx, "/api/entries/exists")
	if err != nil {
		return ctx, err
	}
	existenceURL, err := url.Parse(existenceEndpoint)
	if err != nil {
		return ctx, err
	}
	existenceURL.RawQuery = url.Values{
		"url": []string{pageURL},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, existenceURL.String(), http.NoBody)
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Authorization", tokenResponse.authHeaderValue())
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}
	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("received non-200 status code when checking for entry existence: %d, body: %s", resp.StatusCode, string(body))
	}
	var existenceResult struct {
		Exists bool `json:"exists"`
	}
	err = json.Unmarshal(body, &existenceResult)
	if err != nil {
		return ctx, err
	}
	switch existence {
	case "should":
		if !existenceResult.Exists {
			return ctx, fmt.Errorf("the page does not exist but it should")
		}
	case "should not":
		if existenceResult.Exists {
			return ctx, fmt.Errorf("the page does exists but it should not")
		}
	default:
		return ctx, fmt.Errorf("bad value for existence: %s", existence)
	}
	return ctx, nil
}

type createdEntryKey struct{}

type createdEntry struct {
	ID          int32  `json:"id"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	RetrievedAt string `json:"retrieved_at"`
}

func givenICreatedAValidEntry(ctx context.Context) (context.Context, error) {
	// Set a valid URL for the entry
	pageURL := "https://en.wikipedia.org/wiki/Behavior-driven_development"

	// Use existing authentication from context
	token, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing token response")
	}

	entryEndpoint, err := makeRequestURL(ctx, "/api/entries")
	if err != nil {
		return ctx, err
	}

	formBody := strings.NewReader(url.Values{
		"url": []string{pageURL},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, entryEndpoint, formBody)
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", token.authHeaderValue())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("failed to create entry: status %d, body: %s", resp.StatusCode, string(body))
	}

	var entry createdEntry
	err = json.Unmarshal(body, &entry)
	if err != nil {
		return ctx, fmt.Errorf("failed to parse entry: %v, body: %s", err, string(body))
	}

	return context.WithValue(ctx, createdEntryKey{}, entry), nil
}

type viewEntryResponseKey struct{}

type viewEntryResult struct {
	StatusCode int
	Body       []byte
	Entry      *createdEntry
}

func whenIViewTheCreatedEntry(ctx context.Context) (context.Context, error) {
	entry, ok := ctx.Value(createdEntryKey{}).(createdEntry)
	if !ok {
		return ctx, fmt.Errorf("context is missing created entry")
	}

	token, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("context is missing token response")
	}

	entryEndpoint, err := makeRequestURL(ctx, fmt.Sprintf("/api/entries/%d", entry.ID))
	if err != nil {
		return ctx, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, entryEndpoint, http.NoBody)
	if err != nil {
		return ctx, err
	}
	req.Header.Set("Authorization", token.authHeaderValue())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, err
	}

	result := viewEntryResult{
		StatusCode: resp.StatusCode,
		Body:       body,
	}

	if resp.StatusCode == http.StatusOK {
		var viewedEntry createdEntry
		err = json.Unmarshal(body, &viewedEntry)
		if err != nil {
			return ctx, fmt.Errorf("failed to parse viewed entry: %v", err)
		}
		result.Entry = &viewedEntry
	}

	return context.WithValue(ctx, viewEntryResponseKey{}, result), nil
}

func thenEntryContentShouldMatchTheContentAtCreationTime(ctx context.Context) (context.Context, error) {
	createdEntry, ok := ctx.Value(createdEntryKey{}).(createdEntry)
	if !ok {
		return ctx, fmt.Errorf("context is missing created entry")
	}

	viewResult, ok := ctx.Value(viewEntryResponseKey{}).(viewEntryResult)
	if !ok {
		return ctx, fmt.Errorf("context is missing view entry result")
	}

	if viewResult.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("failed to view entry: status %d, body: %s", viewResult.StatusCode, string(viewResult.Body))
	}

	if viewResult.Entry == nil {
		return ctx, fmt.Errorf("viewed entry is nil")
	}

	if viewResult.Entry.Content != createdEntry.Content {
		return ctx, fmt.Errorf("content mismatch: created=%q, viewed=%q", createdEntry.Content, viewResult.Entry.Content)
	}

	if viewResult.Entry.Title != createdEntry.Title {
		return ctx, fmt.Errorf("title mismatch: created=%q, viewed=%q", createdEntry.Title, viewResult.Entry.Title)
	}

	if viewResult.Entry.URL != createdEntry.URL {
		return ctx, fmt.Errorf("URL mismatch: created=%q, viewed=%q", createdEntry.URL, viewResult.Entry.URL)
	}

	return ctx, nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Given(`there exists (a|my) user account`, givenThereExistsMyUserAccount)
	ctx.Given(`there exists another user account`, givenThereExistsAnotherUserAccount)
	ctx.Given(`there exists a client`, givenThereExistsAClient)
	ctx.Given(`client credentials are (valid|invalid)`, givenClientCredentialsValidity)
	ctx.Given(`user credentials are (valid|invalid)`, givenUserCredentialsValidity)
	ctx.Given(`there is a default client bootstrapped`, givenThereIsADefaultClientBoootstrapped)
	ctx.Given(`there is an admin account bootstrapped`, givenThereIsAnAdminAccountBootstrapped)
	ctx.Given(`bootstrap account credentials are valid`, givenBootstrapAccountCredentialsAreValid)
	ctx.Given(`I am authenticated as admin`, givenIAmAuthenticatedAsAdmin)
	ctx.Given(`entry url points to (valid|invalid) html page`, givenEntryURLPointsToHTMLPage)
	ctx.Given(`I am authenticated as user`, givenIAmAuthenticatedAsUser)
	ctx.Given(`I created a valid entry`, givenICreatedAValidEntry)

	ctx.When(`client uses credentials to authenticate`, whenClientUsesCredentialsToAuthenticate)
	ctx.When(`I use bootstrap credentials to authenticate`, whenIUseBootstrapCredentialsToAuthenticate)
	ctx.When(`I create a new (user|admin) account`, whenICreateANewAccount)
	ctx.When(`I try to add an entry`, whenITryToAddAnEntry)
	ctx.When(`I view the created entry`, whenIViewTheCreatedEntry)

	ctx.Then(`the client should be (authenticated|rejected)`, thenTheClientAuthOutcomeShouldBe)
	ctx.Then(`I am successfully authenticated as admin`, thenIAmSuccessfullyAuthenticatedAsAdmin)
	ctx.Then(`the account creation should succeed`, thenTheAccountCreationShouldSucceed)
	ctx.Then(`I am able to login as that account`, thenIAmAbleToLoginAsThatAccount)
	ctx.Then(`the entry (should|should not) exist`, thenTheEntryShouldHaveExistence)
	ctx.Then(`entry addition (should|should not) succeed`, thenEntryAdditionShouldHaveStatus)
	ctx.Then(`entry content should match the content at creation time`, thenEntryContentShouldMatchTheContentAtCreationTime)
}
