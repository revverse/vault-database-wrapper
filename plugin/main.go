package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	dbplugin "github.com/hashicorp/vault/sdk/database/dbplugin/v5"
)

type MyDBPlugin struct {
    connectionURL string
    username      string
    password      string
    apiBaseURL    string // Base URL for the external service
}
var (
    defaultApiBaseURL="http://172.17.0.1:8888"  // Replace with actual base URL
)
func New() (interface{}, error) {
    db := new()
    dbType := dbplugin.NewDatabaseErrorSanitizerMiddleware(db, db.secretValues)
    return dbType, nil
}

func new() *MyDBPlugin {
    return &MyDBPlugin{
    }
}

func (db *MyDBPlugin) Type() (string, error) {
    return "mydb", nil
}


// Close implements the Close method required by the dbplugin.Database interface.
// Since this plugin uses HTTP calls and doesn't maintain persistent connections,
// this method simply returns nil.
func (db *MyDBPlugin) Close() error {
    return nil
}

func (db *MyDBPlugin) Initialize(ctx context.Context, req dbplugin.InitializeRequest) (dbplugin.InitializeResponse, error) {
    connectionURL, ok := req.Config["connection_url"].(string)
    log.Printf("req", req)
    if !ok {
        return dbplugin.InitializeResponse{}, errors.New("connection_url must be a string")
    }
    username, ok := req.Config["username"].(string)
    if !ok {
        return dbplugin.InitializeResponse{}, errors.New("username must be a string")
    }
    password, ok := req.Config["password"].(string)
    if !ok {
        return dbplugin.InitializeResponse{}, errors.New("password must be a string")
    }
    apiBaseURL, ok := req.Config["api_base_url"].(string)
    if !ok {
        apiBaseURL = defaultApiBaseURL // Default value
    }

    db.connectionURL = connectionURL
    db.username = username
    db.password = password
    db.apiBaseURL = apiBaseURL

    return dbplugin.InitializeResponse{
        Config: req.Config,
    }, nil

}

func (db *MyDBPlugin) NewUser(ctx context.Context, req dbplugin.NewUserRequest) (dbplugin.NewUserResponse, error) {
    expiration := req.Expiration
    if expiration.IsZero() {
        expiration = time.Now().Add(24 * time.Hour) // Default expiration
    }




    payload := map[string]interface{}{
        "username":   req.UsernameConfig.DisplayName, // Use DisplayName from UsernameConfig
        "role_name":  req.UsernameConfig.RoleName,    // Use RoleName from UsernameConfig
        "password":   req.Password,
        "expiration": expiration.Format(time.RFC3339),
        "commands":   req.Statements.Commands,        // Use Commands from Statements
    }

    err := db.makeHTTPRequest(ctx, "POST", "/userAdd", payload)
    if err != nil {
        return dbplugin.NewUserResponse{}, err
    }

    return dbplugin.NewUserResponse{
        Username: req.UsernameConfig.DisplayName, // Return the DisplayName as the username
    }, nil
}

func (db *MyDBPlugin) UpdateUser(ctx context.Context, req dbplugin.UpdateUserRequest) (dbplugin.UpdateUserResponse, error) {
    payload := map[string]interface{}{
        "username": req.Username, // Use DisplayName from UsernameConfig
        "password": req.Password,
    }

    err := db.makeHTTPRequest(ctx, "POST", "/userUpdate", payload)
    if err != nil {
        return dbplugin.UpdateUserResponse{}, err
    }

    return dbplugin.UpdateUserResponse{}, nil
}

func (db *MyDBPlugin) DeleteUser(ctx context.Context, req dbplugin.DeleteUserRequest) (dbplugin.DeleteUserResponse, error) {
    payload := map[string]interface{}{
        "username": req.Username,
    }

    err := db.makeHTTPRequest(ctx, "POST", "/userDelete", payload)
    if err != nil {
        return dbplugin.DeleteUserResponse{}, err
    }

    return dbplugin.DeleteUserResponse{}, nil
}

func (db *MyDBPlugin) changeUserPassword(ctx context.Context, username string, newPassword string, rotateStatements []string, selfManagedPassword string) error {
    payload := map[string]interface{}{
        "username": username,
        "password": newPassword,
    }

    return db.makeHTTPRequest(ctx, "POST", "/passwordChange", payload)
}

func (db *MyDBPlugin) secretValues() map[string]string {
    return map[string]string{
        db.password: "[password]",
    }
}

func (db *MyDBPlugin) parseStatements(rawStatements []string) []string {
    // For simplicity, return the raw statements as-is
    return rawStatements
}

func (db *MyDBPlugin) makeHTTPRequest(ctx context.Context, method, endpoint string, payload map[string]interface{}) error {
    url := fmt.Sprintf("%s%s", db.apiBaseURL, endpoint)
    
    // Add connection details to the payload

    payload["connection_url"] = db.connectionURL
    payload["admin_username"] = db.username
    payload["admin_password"] = db.password


    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("failed to create HTTP request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(respBody))
    }

    return nil
}