package endpoints

import (
	"encoding/json"
	"time"

	"github.com/monzo/slog"
	"github.com/monzo/terrors"
	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/auth"
	"github.com/chongyangshi/yronwood/types"
)

const authTokenValidity = time.Duration(time.Hour * 2)

func authenticate(req typhon.Request) typhon.Response {
	authenticateRequest, err := req.BodyBytes(true)
	if err != nil {
		slog.Error(req, "Error reading request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	body := types.AuthenticateRequest{}
	err = json.Unmarshal(authenticateRequest, &body)
	if err != nil {
		slog.Error(req, "Error parsing request body: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered handling request", nil)}
	}

	if body.Secret == "" {
		return typhon.Response{Error: terrors.BadRequest("secret_required", "Pre-shared secret required for authentication", nil)}
	}

	authencated, err := doBasicAuth(body.Secret)
	if err != nil {
		slog.Error(req, "Error authenticating: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error encountered authenticating request", nil)}
	}
	if !authencated {
		return typhon.Response{Error: terrors.Forbidden("bad_secret", "Incorrect secret, cannot authenticate", nil)}
	}

	token, err := auth.SignAdminToken(authTokenValidity)
	if err != nil {
		slog.Error(req, "Error signing auth token: %v", err)
		return typhon.Response{Error: terrors.InternalService("", "Error processing authenticating request", nil)}
	}

	return req.Response(types.AuthenticateResponse{
		Token: token,
	})
}
