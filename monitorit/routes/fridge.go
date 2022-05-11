package routes

import (
	"context"
	"strconv"
	"time"

	"github.com/cszatmary/fridge-monitor/monitorit/models"
	"github.com/gofiber/fiber/v2"
)

const viewTimeFormat = "Monday, January 2 2006 15:04:05 MST"

type FridgeHandler struct {
	fm *models.FridgeManager
	tm *models.TemperatureManager
}

func NewFridgeHandler(fm *models.FridgeManager, tm *models.TemperatureManager) *FridgeHandler {
	return &FridgeHandler{fm, tm}
}

type fridgeResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MinTemp     float64 `json:"minTemp"`
	MaxTemp     float64 `json:"maxTemp"`
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
			ID:          strconv.FormatInt(f.ID, 10),
			Name:        f.Name,
			Description: f.Description,
			MinTemp:     f.MinTemp,
			MaxTemp:     f.MaxTemp,
		}
	}
	return body, nil
}

type temperatureResponse struct {
	ID        string  `json:"id"`
	Value     float64 `json:"value"`
	CreatedAt string  `json:"createdAt"`
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
			ID:          strconv.FormatInt(fridge.ID, 10),
			Name:        fridge.Name,
			Description: fridge.Description,
			MinTemp:     fridge.MinTemp,
			MaxTemp:     fridge.MaxTemp,
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
				CreatedAt: t.CreatedAt.Local().Format(viewTimeFormat),
			})
		}
	}
	return body, nil
}

func (fh *FridgeHandler) CreateTemperature(ctx context.Context, c *fiber.Ctx) (any, error) {
	fridgeID, err := paramInt64(c, "fridgeID")
	if err != nil {
		return nil, err
	}
	var reqBody struct {
		Value float64 `json:"value"`
	}
	if err := c.BodyParser(&reqBody); err != nil {
		return nil, err
	}
	temp, err := fh.tm.InsertOne(ctx, fridgeID, reqBody.Value)
	if err != nil {
		return nil, err
	}
	return temperatureResponse{
		ID:        strconv.FormatInt(temp.ID, 10),
		Value:     temp.Value,
		CreatedAt: temp.CreatedAt.Format(time.RFC3339),
	}, nil
}
