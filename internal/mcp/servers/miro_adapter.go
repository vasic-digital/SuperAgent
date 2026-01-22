// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// MiroConfig contains configuration for Miro adapter.
type MiroConfig struct {
	AccessToken string        `json:"access_token"`
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
}

// MiroBoard represents a Miro board.
type MiroBoard struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	CreatedAt       time.Time          `json:"createdAt"`
	ModifiedAt      time.Time          `json:"modifiedAt"`
	ViewLink        string             `json:"viewLink"`
	AccessLink      string             `json:"accessLink"`
	Picture         *MiroPicture       `json:"picture,omitempty"`
	Team            *MiroTeam          `json:"team,omitempty"`
	Owner           *MiroUser          `json:"owner,omitempty"`
	CurrentUserRole string             `json:"currentUserMembership,omitempty"`
	Policy          *MiroBoardPolicy   `json:"policy,omitempty"`
	SharingPolicy   *MiroSharingPolicy `json:"sharingPolicy,omitempty"`
}

// MiroPicture represents a board picture/thumbnail.
type MiroPicture struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageURL"`
}

// MiroTeam represents a Miro team.
type MiroTeam struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MiroUser represents a Miro user.
type MiroUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// MiroBoardPolicy represents board access policies.
type MiroBoardPolicy struct {
	PermissionsPolicy *MiroPermissionsPolicy `json:"permissionsPolicy,omitempty"`
	SharingPolicy     *MiroSharingPolicy     `json:"sharingPolicy,omitempty"`
}

// MiroPermissionsPolicy defines permissions for a board.
type MiroPermissionsPolicy struct {
	CollaborationToolsStartAccess string `json:"collaborationToolsStartAccess"`
	CopyAccess                    string `json:"copyAccess"`
	SharingAccess                 string `json:"sharingAccess"`
}

// MiroSharingPolicy defines sharing settings.
type MiroSharingPolicy struct {
	Access                      string `json:"access"`
	InviteToAccountAndBoardLink string `json:"inviteToAccountAndBoardLink"`
	OrganizationAccess          string `json:"organizationAccess"`
	TeamAccess                  string `json:"teamAccess"`
}

// MiroItem represents a generic item on a board.
type MiroItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Position   *MiroPosition          `json:"position,omitempty"`
	Geometry   *MiroGeometry          `json:"geometry,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Style      map[string]interface{} `json:"style,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
	ModifiedAt time.Time              `json:"modifiedAt"`
	CreatedBy  *MiroUser              `json:"createdBy,omitempty"`
	ModifiedBy *MiroUser              `json:"modifiedBy,omitempty"`
	Parent     *MiroItemRef           `json:"parent,omitempty"`
}

// MiroPosition represents item position on canvas.
type MiroPosition struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Origin string  `json:"origin,omitempty"`
}

// MiroGeometry represents item dimensions.
type MiroGeometry struct {
	Width    float64 `json:"width,omitempty"`
	Height   float64 `json:"height,omitempty"`
	Rotation float64 `json:"rotation,omitempty"`
}

// MiroItemRef is a reference to another item.
type MiroItemRef struct {
	ID string `json:"id"`
}

// MiroStickyNote represents a sticky note item.
type MiroStickyNote struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Data     *MiroStickyNoteData  `json:"data"`
	Style    *MiroStickyNoteStyle `json:"style,omitempty"`
	Position *MiroPosition        `json:"position,omitempty"`
	Geometry *MiroGeometry        `json:"geometry,omitempty"`
}

// MiroStickyNoteData contains sticky note content.
type MiroStickyNoteData struct {
	Content string `json:"content"`
	Shape   string `json:"shape,omitempty"`
}

// MiroStickyNoteStyle contains sticky note styling.
type MiroStickyNoteStyle struct {
	FillColor         string `json:"fillColor,omitempty"`
	TextAlign         string `json:"textAlign,omitempty"`
	TextAlignVertical string `json:"textAlignVertical,omitempty"`
}

// MiroShape represents a shape item.
type MiroShape struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Data     *MiroShapeData  `json:"data"`
	Style    *MiroShapeStyle `json:"style,omitempty"`
	Position *MiroPosition   `json:"position,omitempty"`
	Geometry *MiroGeometry   `json:"geometry,omitempty"`
}

// MiroShapeData contains shape content.
type MiroShapeData struct {
	Content string `json:"content,omitempty"`
	Shape   string `json:"shape"`
}

// MiroShapeStyle contains shape styling.
type MiroShapeStyle struct {
	BorderColor   string  `json:"borderColor,omitempty"`
	BorderOpacity float64 `json:"borderOpacity,omitempty"`
	BorderStyle   string  `json:"borderStyle,omitempty"`
	BorderWidth   float64 `json:"borderWidth,omitempty"`
	Color         string  `json:"color,omitempty"`
	FillColor     string  `json:"fillColor,omitempty"`
	FillOpacity   float64 `json:"fillOpacity,omitempty"`
	FontFamily    string  `json:"fontFamily,omitempty"`
	FontSize      int     `json:"fontSize,omitempty"`
	TextAlign     string  `json:"textAlign,omitempty"`
}

// MiroConnector represents a connector/line between items.
type MiroConnector struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	StartItem  *MiroConnectorEnd   `json:"startItem,omitempty"`
	EndItem    *MiroConnectorEnd   `json:"endItem,omitempty"`
	Style      *MiroConnectorStyle `json:"style,omitempty"`
	Captions   []MiroCaption       `json:"captions,omitempty"`
	Shape      string              `json:"shape,omitempty"`
	CreatedAt  time.Time           `json:"createdAt"`
	ModifiedAt time.Time           `json:"modifiedAt"`
}

// MiroConnectorEnd represents an endpoint of a connector.
type MiroConnectorEnd struct {
	ID       string        `json:"id"`
	Position *MiroPosition `json:"position,omitempty"`
	SnapTo   string        `json:"snapTo,omitempty"`
}

// MiroConnectorStyle contains connector styling.
type MiroConnectorStyle struct {
	Color           string  `json:"color,omitempty"`
	EndStrokeCap    string  `json:"endStrokeCap,omitempty"`
	FontFamily      string  `json:"fontFamily,omitempty"`
	FontSize        int     `json:"fontSize,omitempty"`
	StartStrokeCap  string  `json:"startStrokeCap,omitempty"`
	StrokeColor     string  `json:"strokeColor,omitempty"`
	StrokeStyle     string  `json:"strokeStyle,omitempty"`
	StrokeWidth     float64 `json:"strokeWidth,omitempty"`
	TextOrientation string  `json:"textOrientation,omitempty"`
}

// MiroCaption represents a caption on a connector.
type MiroCaption struct {
	Content  string        `json:"content"`
	Position *MiroPosition `json:"position,omitempty"`
}

// MiroFrame represents a frame on the board.
type MiroFrame struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Data     *MiroFrameData  `json:"data"`
	Style    *MiroFrameStyle `json:"style,omitempty"`
	Position *MiroPosition   `json:"position,omitempty"`
	Geometry *MiroGeometry   `json:"geometry,omitempty"`
	Children []string        `json:"children,omitempty"`
}

// MiroFrameData contains frame metadata.
type MiroFrameData struct {
	Title  string `json:"title"`
	Format string `json:"format,omitempty"`
	Type   string `json:"type,omitempty"`
}

// MiroFrameStyle contains frame styling.
type MiroFrameStyle struct {
	FillColor string `json:"fillColor,omitempty"`
}

// MiroText represents a text item.
type MiroText struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Data     *MiroTextData  `json:"data"`
	Style    *MiroTextStyle `json:"style,omitempty"`
	Position *MiroPosition  `json:"position,omitempty"`
	Geometry *MiroGeometry  `json:"geometry,omitempty"`
}

// MiroTextData contains text content.
type MiroTextData struct {
	Content string `json:"content"`
}

// MiroTextStyle contains text styling.
type MiroTextStyle struct {
	Color       string  `json:"color,omitempty"`
	FillColor   string  `json:"fillColor,omitempty"`
	FillOpacity float64 `json:"fillOpacity,omitempty"`
	FontFamily  string  `json:"fontFamily,omitempty"`
	FontSize    int     `json:"fontSize,omitempty"`
	TextAlign   string  `json:"textAlign,omitempty"`
}

// MiroImage represents an image item.
type MiroImage struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Data     *MiroImageData `json:"data"`
	Position *MiroPosition  `json:"position,omitempty"`
	Geometry *MiroGeometry  `json:"geometry,omitempty"`
}

// MiroImageData contains image metadata.
type MiroImageData struct {
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title,omitempty"`
}

// MiroCard represents a card item.
type MiroCard struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Data     *MiroCardData  `json:"data"`
	Style    *MiroCardStyle `json:"style,omitempty"`
	Position *MiroPosition  `json:"position,omitempty"`
	Geometry *MiroGeometry  `json:"geometry,omitempty"`
}

// MiroCardData contains card content.
type MiroCardData struct {
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	Assignee    *MiroUser  `json:"assignee,omitempty"`
}

// MiroCardStyle contains card styling.
type MiroCardStyle struct {
	CardTheme string `json:"cardTheme,omitempty"`
}

// MiroAdapter implements ServerAdapter for Miro API.
type MiroAdapter struct {
	mu        sync.RWMutex
	config    MiroConfig
	client    *http.Client
	connected bool
	baseURL   string
}

// NewMiroAdapter creates a new Miro adapter.
func NewMiroAdapter(config MiroConfig) *MiroAdapter {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.miro.com/v2"
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &MiroAdapter{
		config:  config,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Connect establishes connection to Miro API.
func (a *MiroAdapter) Connect(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Verify token by getting current user
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/users/me", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Miro: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid access token")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to authenticate: %s", string(body))
	}

	a.connected = true
	return nil
}

// Close closes the adapter connection.
func (a *MiroAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	return nil
}

// Health checks if the adapter is healthy.
func (a *MiroAdapter) Health(ctx context.Context) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected {
		return fmt.Errorf("not connected")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/users/me", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// ListBoards retrieves all boards accessible to the user.
func (a *MiroAdapter) ListBoards(ctx context.Context, teamID string, limit int, cursor string) ([]MiroBoard, string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	params := url.Values{}
	if teamID != "" {
		params.Set("team_id", teamID)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	endpoint := a.baseURL + "/boards"
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list boards: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to list boards: %s", string(body))
	}

	var result struct {
		Data   []MiroBoard `json:"data"`
		Cursor string      `json:"cursor,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, result.Cursor, nil
}

// GetBoard retrieves a specific board by ID.
func (a *MiroAdapter) GetBoard(ctx context.Context, boardID string) (*MiroBoard, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/boards/"+boardID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("board not found: %s", boardID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get board: %s", string(body))
	}

	var board MiroBoard
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode board: %w", err)
	}

	return &board, nil
}

// CreateBoard creates a new board.
func (a *MiroAdapter) CreateBoard(ctx context.Context, name, description string, teamID string) (*MiroBoard, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"name": name,
	}
	if description != "" {
		payload["description"] = description
	}
	if teamID != "" {
		payload["teamId"] = teamID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/boards", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create board: %s", string(respBody))
	}

	var board MiroBoard
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode board: %w", err)
	}

	return &board, nil
}

// UpdateBoard updates board properties.
func (a *MiroAdapter) UpdateBoard(ctx context.Context, boardID, name, description string) (*MiroBoard, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{}
	if name != "" {
		payload["name"] = name
	}
	if description != "" {
		payload["description"] = description
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", a.baseURL+"/boards/"+boardID, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update board: %s", string(respBody))
	}

	var board MiroBoard
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode board: %w", err)
	}

	return &board, nil
}

// DeleteBoard deletes a board.
func (a *MiroAdapter) DeleteBoard(ctx context.Context, boardID string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "DELETE", a.baseURL+"/boards/"+boardID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete board: %s", string(body))
	}

	return nil
}

// GetBoardItems retrieves all items from a board.
func (a *MiroAdapter) GetBoardItems(ctx context.Context, boardID string, itemType string, limit int, cursor string) ([]MiroItem, string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	params := url.Values{}
	if itemType != "" {
		params.Set("type", itemType)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/items", a.baseURL, boardID)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get items: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to get items: %s", string(body))
	}

	var result struct {
		Data   []MiroItem `json:"data"`
		Cursor string     `json:"cursor,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Data, result.Cursor, nil
}

// GetItem retrieves a specific item by ID.
func (a *MiroAdapter) GetItem(ctx context.Context, boardID, itemID string) (*MiroItem, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/boards/%s/items/%s", a.baseURL, boardID, itemID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("item not found: %s", itemID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get item: %s", string(body))
	}

	var item MiroItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to decode item: %w", err)
	}

	return &item, nil
}

// CreateStickyNote creates a sticky note on a board.
func (a *MiroAdapter) CreateStickyNote(ctx context.Context, boardID string, content string, position *MiroPosition, style *MiroStickyNoteStyle) (*MiroStickyNote, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"content": content,
			"shape":   "square",
		},
	}
	if position != nil {
		payload["position"] = position
	}
	if style != nil {
		payload["style"] = style
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/sticky_notes", a.baseURL, boardID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create sticky note: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create sticky note: %s", string(respBody))
	}

	var note MiroStickyNote
	if err := json.NewDecoder(resp.Body).Decode(&note); err != nil {
		return nil, fmt.Errorf("failed to decode sticky note: %w", err)
	}

	return &note, nil
}

// CreateShape creates a shape on a board.
func (a *MiroAdapter) CreateShape(ctx context.Context, boardID string, shapeType string, content string, position *MiroPosition, geometry *MiroGeometry, style *MiroShapeStyle) (*MiroShape, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	data := map[string]interface{}{
		"shape": shapeType,
	}
	if content != "" {
		data["content"] = content
	}

	payload := map[string]interface{}{
		"data": data,
	}
	if position != nil {
		payload["position"] = position
	}
	if geometry != nil {
		payload["geometry"] = geometry
	}
	if style != nil {
		payload["style"] = style
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/shapes", a.baseURL, boardID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create shape: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create shape: %s", string(respBody))
	}

	var shape MiroShape
	if err := json.NewDecoder(resp.Body).Decode(&shape); err != nil {
		return nil, fmt.Errorf("failed to decode shape: %w", err)
	}

	return &shape, nil
}

// CreateConnector creates a connector between two items.
func (a *MiroAdapter) CreateConnector(ctx context.Context, boardID string, startItemID, endItemID string, style *MiroConnectorStyle) (*MiroConnector, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"startItem": map[string]interface{}{
			"id": startItemID,
		},
		"endItem": map[string]interface{}{
			"id": endItemID,
		},
	}
	if style != nil {
		payload["style"] = style
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/connectors", a.baseURL, boardID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create connector: %s", string(respBody))
	}

	var connector MiroConnector
	if err := json.NewDecoder(resp.Body).Decode(&connector); err != nil {
		return nil, fmt.Errorf("failed to decode connector: %w", err)
	}

	return &connector, nil
}

// CreateFrame creates a frame on a board.
func (a *MiroAdapter) CreateFrame(ctx context.Context, boardID string, title string, position *MiroPosition, geometry *MiroGeometry, style *MiroFrameStyle) (*MiroFrame, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"title":  title,
			"format": "custom",
		},
	}
	if position != nil {
		payload["position"] = position
	}
	if geometry != nil {
		payload["geometry"] = geometry
	}
	if style != nil {
		payload["style"] = style
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/frames", a.baseURL, boardID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create frame: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create frame: %s", string(respBody))
	}

	var frame MiroFrame
	if err := json.NewDecoder(resp.Body).Decode(&frame); err != nil {
		return nil, fmt.Errorf("failed to decode frame: %w", err)
	}

	return &frame, nil
}

// CreateText creates a text item on a board.
func (a *MiroAdapter) CreateText(ctx context.Context, boardID string, content string, position *MiroPosition, style *MiroTextStyle) (*MiroText, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"content": content,
		},
	}
	if position != nil {
		payload["position"] = position
	}
	if style != nil {
		payload["style"] = style
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/boards/%s/texts", a.baseURL, boardID)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create text: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create text: %s", string(respBody))
	}

	var text MiroText
	if err := json.NewDecoder(resp.Body).Decode(&text); err != nil {
		return nil, fmt.Errorf("failed to decode text: %w", err)
	}

	return &text, nil
}

// DeleteItem deletes an item from a board.
func (a *MiroAdapter) DeleteItem(ctx context.Context, boardID, itemID string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	endpoint := fmt.Sprintf("%s/boards/%s/items/%s", a.baseURL, boardID, itemID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.config.AccessToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete item: %s", string(body))
	}

	return nil
}

// GetMCPTools returns the MCP tool definitions for Miro.
func (a *MiroAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "miro_list_boards",
			Description: "List all Miro boards accessible to the user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by team ID",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of boards to return",
					},
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "Pagination cursor",
					},
				},
			},
		},
		{
			Name:        "miro_get_board",
			Description: "Get details of a specific Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "miro_create_board",
			Description: "Create a new Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Board name",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Board description",
					},
					"team_id": map[string]interface{}{
						"type":        "string",
						"description": "Team ID to create the board in",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "miro_delete_board",
			Description: "Delete a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID to delete",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "miro_get_board_items",
			Description: "Get all items from a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by item type (sticky_note, shape, text, etc.)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of items to return",
					},
					"cursor": map[string]interface{}{
						"type":        "string",
						"description": "Pagination cursor",
					},
				},
				"required": []string{"board_id"},
			},
		},
		{
			Name:        "miro_create_sticky_note",
			Description: "Create a sticky note on a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Sticky note content",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
					},
					"fill_color": map[string]interface{}{
						"type":        "string",
						"description": "Fill color (e.g., yellow, cyan, green)",
					},
				},
				"required": []string{"board_id", "content"},
			},
		},
		{
			Name:        "miro_create_shape",
			Description: "Create a shape on a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"shape": map[string]interface{}{
						"type":        "string",
						"description": "Shape type (rectangle, circle, triangle, etc.)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Shape text content",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width",
					},
					"height": map[string]interface{}{
						"type":        "number",
						"description": "Height",
					},
				},
				"required": []string{"board_id", "shape"},
			},
		},
		{
			Name:        "miro_create_connector",
			Description: "Create a connector between two items on a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"start_item_id": map[string]interface{}{
						"type":        "string",
						"description": "Starting item ID",
					},
					"end_item_id": map[string]interface{}{
						"type":        "string",
						"description": "Ending item ID",
					},
				},
				"required": []string{"board_id", "start_item_id", "end_item_id"},
			},
		},
		{
			Name:        "miro_create_frame",
			Description: "Create a frame on a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Frame title",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
					},
					"width": map[string]interface{}{
						"type":        "number",
						"description": "Width",
					},
					"height": map[string]interface{}{
						"type":        "number",
						"description": "Height",
					},
				},
				"required": []string{"board_id", "title"},
			},
		},
		{
			Name:        "miro_create_text",
			Description: "Create a text item on a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Text content",
					},
					"x": map[string]interface{}{
						"type":        "number",
						"description": "X position",
					},
					"y": map[string]interface{}{
						"type":        "number",
						"description": "Y position",
					},
					"font_size": map[string]interface{}{
						"type":        "integer",
						"description": "Font size",
					},
				},
				"required": []string{"board_id", "content"},
			},
		},
		{
			Name:        "miro_delete_item",
			Description: "Delete an item from a Miro board",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "The board ID",
					},
					"item_id": map[string]interface{}{
						"type":        "string",
						"description": "The item ID to delete",
					},
				},
				"required": []string{"board_id", "item_id"},
			},
		},
	}
}
