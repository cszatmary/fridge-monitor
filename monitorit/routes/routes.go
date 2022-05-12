package routes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/cszatmary/fridge-monitor/monitorit/lib/apierror"
	"github.com/cszatmary/fridge-monitor/monitorit/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/html"
)

// Set during docker build
var gitsha = "unavailable"

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func SetupApp(db *sql.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		Views:       html.New("./resources/views", ".gohtml"),
		ViewsLayout: "layouts/page",
		AppName:     "MonitorIt",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := apierror.CodeUnknown
			errorResp := errorResponse{
				Code:    code.String(),
				Message: "An unknown error occurred",
			}

			var apiErr apierror.Error
			if errors.As(err, &apiErr) {
				code = apiErr.Code()
				errorResp.Code = code.String()
				errorResp.Message = apiErr.Message()
			}

			var detailedErr apierror.DetailedError
			if errors.As(err, &detailedErr) {
				log.Printf("Error: %s", detailedErr.Details())
			}

			// Default status code to 500
			status := fiber.StatusInternalServerError
			switch code {
			case apierror.CodeDatabase:
				status = fiber.StatusInternalServerError
			case apierror.CodeRecordNotFound:
				status = fiber.StatusNotFound
			case apierror.CodeInvalidParameter:
				status = fiber.StatusBadRequest
			}

			body := struct {
				Error  errorResponse `json:"error"`
				Status int           `json:"-"`
			}{Error: errorResp, Status: status}
			if isHTML(c) {
				return c.Render("error", body)
			}
			return c.JSON(body)
		},
	})
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recovermw.New())

	fm := models.NewFridgeManager(db)
	tm := models.NewTemperatureManager(db)
	fh := NewFridgeHandler(fm, tm)

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("MonitorIt OK: " + gitsha)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/fridges")
	})
	app.Get("/fridges", createHandler("fridges/index", fh.List))
	app.Get("/fridges/:fridgeID", createHandler("fridges/show", fh.Get))
	app.Post("/fridges", createHandler("", withTransaction(db, fh.Create)))
	app.Post("/fridges/:fridgeID/temperatures", createHandler("", withTransaction(db, fh.CreateTemperature)))
	return app
}

type handler func(context.Context, *fiber.Ctx) (any, error)

func createHandler(templateName string, h handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := h(c.Context(), c)
		if err != nil {
			return err
		}
		if templateName != "" && isHTML(c) {
			return c.Render(templateName, data)
		}
		return c.JSON(data)
	}
}

func withTransaction(db *sql.DB, h handler) handler {
	return handler(func(ctx context.Context, c *fiber.Ctx) (data any, err error) {
		const op = apierror.Op("routes.withTransaction")
		txn, err := db.Begin()
		if err != nil {
			return nil, apierror.Wrap(
				err,
				apierror.CodeDatabase,
				"failed to create database transaction",
				op,
			)
		}

		// Add the txn to the context so it can be used by handlers
		ctx = models.ContextWithTxn(ctx, txn)

		// Handle end of request in a defer
		// That way we can easily handle success, failure, and panic in one place
		defer func() {
			// Create recoverer so we can rollback if a panic occurs
			if v := recover(); v != nil {
				// A panic occurred, capture the error so we can rollback
				// and trigger the error handler
				switch v := v.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%v", v)
				}
			}
			if err != nil {
				// Either an error occurred in the handler or a panic occurred
				// and was recovered above
				// In either case we need to rollback the txn and call the error handler
				rollbackErr := txn.Rollback()
				if rollbackErr != nil {
					log.Printf("Failed to rollback database transaction: %v", rollbackErr)
				}
				return
			}
			// No error occurred, we are good to commit the txn
			commitErr := txn.Commit()
			if commitErr != nil {
				log.Printf("Failed to commit database transaction: %v", commitErr)
			}
		}()

		// Continue handling the request
		return h(ctx, c)
	})
}

func isHTML(c *fiber.Ctx) bool {
	return c.Accepts("application/json", "text/html") == "text/html"
}

func paramInt64(c *fiber.Ctx, key string) (int64, error) {
	raw := c.Params(key)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, apierror.Wrap(
			err,
			apierror.CodeInvalidParameter,
			fmt.Sprintf("failed to parse %s %q", key, raw),
			"routes.paramInt64",
		)
	}
	return v, nil
}
