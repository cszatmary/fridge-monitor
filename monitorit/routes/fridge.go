package routes

import (
	"context"
	"strconv"
	"time"

	"github.com/cszatmary/fridge-monitor/monitorit/models"
	"github.com/gofiber/fiber/v2"
)

type FridgeHandler struct {
	fm *models.FridgeManager
	tm *models.TemperatureManager
}

func NewFridgeHandler(fm *models.FridgeManager, tm *models.TemperatureManager) *FridgeHandler {
	return &FridgeHandler{fm, tm}
}

type fridgeResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	MinTemp       float64 `json:"minTemp"`
	MaxTemp       float64 `json:"maxTemp"`
	AlertsEnabled bool    `json:"alertsEnabled"`
}

func (fh *FridgeHandler) List(ctx context.Context, c *fiber.Ctx) (any, error) {
	fridges, err := fh.fm.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	body := struct {
		Fridges []fridgeResponse `json:"fridges"`
	}{Fridges: make([]fridgeResponse, len(fridges))}
	for i, f := range fridges {
		body.Fridges[i] = fridgeResponse{
			ID:            strconv.FormatInt(f.ID, 10),
			Name:          f.Name,
			Description:   f.Description,
			MinTemp:       f.MinTemp,
			MaxTemp:       f.MaxTemp,
			AlertsEnabled: f.AlertsEnabled,
		}
	}
	return body, nil
}

type temperatureResponse struct {
	ID        string  `json:"id"`
	Value     float64 `json:"value"`
	Humidity  float64 `json:"humidity"`
	CreatedAt string  `json:"createdAt"`
	Status    string  `json:"-"`
}

func (fh *FridgeHandler) Get(ctx context.Context, c *fiber.Ctx) (any, error) {
	id, err := paramInt64(c, "fridgeID")
	if err != nil {
		return nil, err
	}
	fridge, err := fh.fm.FindOneByID(ctx, id)
	if err != nil {
		return nil, err
	}

	body := struct {
		fridgeResponse
		Temperatures []temperatureResponse `json:"temperatures,omitempty"`
	}{
		fridgeResponse: fridgeResponse{
			ID:            strconv.FormatInt(fridge.ID, 10),
			Name:          fridge.Name,
			Description:   fridge.Description,
			MinTemp:       fridge.MinTemp,
			MaxTemp:       fridge.MaxTemp,
			AlertsEnabled: fridge.AlertsEnabled,
		},
	}
	if isHTML(c) {
		// If html then also include the last 5 temperatures to display in the view
		temperatures, err := fh.tm.FindMostRecentByFridgeID(ctx, fridge.ID, 5)
		if err != nil {
			return nil, err
		}
		for _, t := range temperatures {
			body.Temperatures = append(body.Temperatures, temperatureResponse{
				ID:        strconv.FormatInt(t.ID, 10),
				Value:     t.Value,
				Humidity:  t.Humidity,
				CreatedAt: t.CreatedAt.Local().Format(models.TimeFormatPretty),
				Status:    t.Status(fridge.MinTemp, fridge.MaxTemp).String(),
			})
		}
	}
	return body, nil
}

func (fh *FridgeHandler) Create(ctx context.Context, c *fiber.Ctx) (any, error) {
	var reqBody fridgeResponse
	if err := c.BodyParser(&reqBody); err != nil {
		return nil, err
	}
	f, err := fh.fm.InsertOne(ctx, models.Fridge{
		Name:          reqBody.Name,
		Description:   reqBody.Description,
		MinTemp:       reqBody.MinTemp,
		MaxTemp:       reqBody.MaxTemp,
		AlertsEnabled: reqBody.AlertsEnabled,
	})
	if err != nil {
		return nil, err
	}
	return fridgeResponse{
		ID:            strconv.FormatInt(f.ID, 10),
		Name:          f.Name,
		Description:   f.Description,
		MinTemp:       f.MinTemp,
		MaxTemp:       f.MaxTemp,
		AlertsEnabled: f.AlertsEnabled,
	}, nil
}

func (fh *FridgeHandler) Update(ctx context.Context, c *fiber.Ctx) (any, error) {
	id, err := paramInt64(c, "fridgeID")
	if err != nil {
		return nil, err
	}
	var reqBody struct {
		Name          string   `json:"name"`
		Description   *string  `json:"description"`
		MinTemp       *float64 `json:"minTemp"`
		MaxTemp       *float64 `json:"maxTemp"`
		AlertsEnabled *bool    `json:"alertsEnabled"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return nil, err
	}

	f, err := fh.fm.UpdateOne(ctx, id, models.PartialFridge{
		Name:          reqBody.Name,
		Description:   reqBody.Description,
		MinTemp:       reqBody.MinTemp,
		MaxTemp:       reqBody.MaxTemp,
		AlertsEnabled: reqBody.AlertsEnabled,
	})
	if err != nil {
		return nil, err
	}
	return fridgeResponse{
		ID:            strconv.FormatInt(f.ID, 10),
		Name:          f.Name,
		Description:   f.Description,
		MinTemp:       f.MinTemp,
		MaxTemp:       f.MaxTemp,
		AlertsEnabled: f.AlertsEnabled,
	}, nil
}

func (fh *FridgeHandler) CreateTemperature(ctx context.Context, c *fiber.Ctx) (any, error) {
	fridgeID, err := paramInt64(c, "fridgeID")
	if err != nil {
		return nil, err
	}
	var reqBody struct {
		Value    float64 `json:"value"`
		Humidity float64 `json:"humidity"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return nil, err
	}
	temp, err := fh.tm.InsertOne(ctx, fridgeID, reqBody.Value, reqBody.Humidity)
	if err != nil {
		return nil, err
	}
	return temperatureResponse{
		ID:        strconv.FormatInt(temp.ID, 10),
		Value:     temp.Value,
		Humidity:  temp.Humidity,
		CreatedAt: temp.CreatedAt.Format(time.RFC3339),
	}, nil
}
