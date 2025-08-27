package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

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

func thenIAmPreventedFromDeletingTheAccount() error {
	return godog.ErrUndefined
}

func thenIAmSuccessfullyAuthenticatedAsAdmin(ctx context.Context) (context.Context, error) {
	_, ok := ctx.Value(tokenResponseKey{}).(tokenResponse)
	if !ok {
		return ctx, fmt.Errorf("unable to obtain token response")
	}
	return ctx, nil
}

func whenICreateANewAccount() error {
	return godog.ErrUndefined
}

func whenITryToDeleteAccount(_ string) error {
	return godog.ErrUndefined
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

func givenAnotherAccountExists() error {
	return godog.ErrUndefined
}

func thenAccountExistenceIsAsExpected() error {
	return godog.ErrUndefined
}

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func givenThereExistsAUserAccount(ctx context.Context) (context.Context, error) {
	// well use the bootstrapped client for now
	return ctx, nil
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

func givenIAmAuthenticated(ctx context.Context) (context.Context, error) {
	// lets use admin account for now
	return givenIAmAuthenticatedAsAdmin(ctx)
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
			return ctx, fmt.Errorf("bad status code, expected 200 but got %d", response.StatusCode)
		}
	case "should not":
		//nolint:usestdlibvars // false-positive for 100 -> http.StatusContinue
		if (response.StatusCode % 100) != 4 {
			return ctx, fmt.Errorf("bad status code, expected 4xx but got %d", response.StatusCode)
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
	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("received non-200 status code when checking for entry existence: %d", resp.StatusCode)
	}
	var existenceResult struct {
		Exists bool `json:"exists"`
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return ctx, err
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

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Given(`there exists a user account`, givenThereExistsAUserAccount)
	ctx.Given(`there exists a client`, givenThereExistsAClient)
	ctx.Given(`client credentials are (valid|invalid)`, givenClientCredentialsValidity)
	ctx.Given(`user credentials are (valid|invalid)`, givenUserCredentialsValidity)
	ctx.Given(`there is a default client bootstrapped`, givenThereIsADefaultClientBoootstrapped)
	ctx.Given(`there is an admin account bootstrapped`, givenThereIsAnAdminAccountBootstrapped)
	ctx.Given(`bootstrap account credentials are valid`, givenBootstrapAccountCredentialsAreValid)
	ctx.Given(`I am authenticated as admin`, givenIAmAuthenticatedAsAdmin)
	ctx.Given(`there exists another (user|admin) account`, givenAnotherAccountExists)
	ctx.Given(`entry url points to (valid|invalid) html page`, givenEntryURLPointsToHTMLPage)
	ctx.Given(`I am authenticated`, givenIAmAuthenticated)

	ctx.When(`client uses credentials to authenticate`, whenClientUsesCredentialsToAuthenticate)
	ctx.When(`I use bootstrap credentials to authenticate`, whenIUseBootstrapCredentialsToAuthenticate)
	ctx.When(`I create a new (user|admin) account`, whenICreateANewAccount)
	ctx.When(`I (?:try to )?delete (my|bootstrapped admin|that) account`, whenITryToDeleteAccount)
	ctx.When(`I try to add an entry`, whenITryToAddAnEntry)

	ctx.Then(`the client should be (authenticated|rejected)`, thenTheClientAuthOutcomeShouldBe)
	ctx.Then(`I am successfully authenticated as admin`, thenIAmSuccessfullyAuthenticatedAsAdmin)
	ctx.Then(`I am prevented from deleting the account`, thenIAmPreventedFromDeletingTheAccount)
	ctx.Then(`(my|bootstrapped admin|admin|user) account (exists|still exists|no longer exists)`, thenAccountExistenceIsAsExpected)
	ctx.Then(`the entry (should|should not) exist`, thenTheEntryShouldHaveExistence)
	ctx.Then(`entry addition (should|should not) succeed`, thenEntryAdditionShouldHaveStatus)
}
