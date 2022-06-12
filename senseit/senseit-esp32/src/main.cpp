#include "display.hpp"
#include "secrets.h"
#include <Adafruit_BME280.h>
#include <Adafruit_Sensor.h>
#include <Arduino.h>
#include <HTTPClient.h>
#include <SPI.h>
#include <WiFi.h>
#include <Wire.h>

#define UPDATE_INTERVAL_MS 30000 // 30sec
#define POST_INTERVAL_MS 600000  // 10min

unsigned long lastTimeUpdate = 0;
unsigned long lastTimePost = 0;
bool retryRequest = false;
Display display;
Adafruit_BME280 sensor;

void handleWiFiStationConnected(WiFiEvent_t event, WiFiEventInfo_t info) {
    Serial.println("Connected to WiFi network " WIFI_SSID " successfully!");
}

void handleWiFiGotIP(WiFiEvent_t event, WiFiEventInfo_t info) {
    Serial.printf("WiFi IP address: %s\n", WiFi.localIP().toString().c_str());
}

void handleWiFiStationDisconnected(WiFiEvent_t event, WiFiEventInfo_t info) {
    Serial.printf("WiFi lost connection. Reason: %u\nTrying to Reconnect\n",
                  info.wifi_sta_disconnected.reason);
    WiFi.disconnect();
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
}

void setup() {
    Serial.begin(115200);
    // Setup the screen. Do this first so that if any subsequent steps fail
    // or any errors occur we can display it on the screen.
    if (!Wire.begin(SCREEN_I2C_SDA, SCREEN_I2C_SCL)) {
        Serial.println();
        // Nothing we can really do here.
        while (true) {
            Serial.println("Screen wire initialization failed");
            delay(10000);
        }
    }
    if (!display.begin(&Wire, SCREEN_I2C_ADDRESS, SCREEN_WIDTH,
                       SCREEN_HEIGHT)) {
        // Nothing we can really do here.
        while (true) {
            Serial.println("Screen setup failed");
            delay(10000);
        }
    }
    display.minSafeTemp = MIN_SAFE_TEMP;
    display.maxSafeTemp = MAX_SAFE_TEMP;
    display.printToSerial = true;

    if (!Wire1.begin(SENSOR_I2C_SDA, SENSOR_I2C_SCL, 100000U)) {
        display.setError(
            {.title = "Setup failed", .message = "Sensor I2C setup failed"});
        display.render();
        while (true) {
            delay(10000);
        }
    }
    if (!sensor.begin(SENSOR_I2C_ADDRESS, &Wire1)) {
        display.setError({.title = "Setup failed",
                          .message = "BME280 sensor initialization failed"});
        display.render();
        while (true) {
            delay(10000);
        }
    }

    // Setup WiFi on boot.
    // Register event handlers so that it automatically reconnects if the
    // connection is lost.
    display.setInfo(
        {.title = "Performing setup", .message = "Connecting to WiFi"});
    WiFi.onEvent(handleWiFiStationConnected,
                 WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_CONNECTED);
    WiFi.onEvent(handleWiFiGotIP, WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_GOT_IP);
    WiFi.onEvent(handleWiFiStationDisconnected,
                 WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_DISCONNECTED);
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    // Wait a bit to connect, if still no connection then give up and move on.
    for (int i = 0; i < 5; i++) {
        if (WiFi.status() == WL_CONNECTED) {
            break;
        }
        delay(500);
    }

    display.clear();
    // Display the temp immediately before beginning the periodic update.
    display.setTemperature({sensor.readTemperature(), sensor.readHumidity()});
    display.render();
}

void loop() {
    if ((millis() - lastTimeUpdate) < UPDATE_INTERVAL_MS) {
        return;
    }
    lastTimeUpdate = millis();
    Temperature temp = {sensor.readTemperature(), sensor.readHumidity()};
    display.setTemperature(temp);
    display.render();

    if (!retryRequest || (millis() - lastTimePost) < POST_INTERVAL_MS) {
        return;
    }
    lastTimePost = millis();

    // Also send a POST request with the temperature to MonitorIt.
    WiFiClient wifiClient;
    HTTPClient http;
    http.begin(wifiClient, MONITORIT_URL);
    http.addHeader("Content-Type", "application/json");
    // This is a little hacky but the JSON body is so simple that it seems
    // overkill to add a JSON library as a dependency just for this.
    // This works well enough.
    String requestData;
    requestData.reserve(40);
    requestData += "{\"value\":";
    requestData += String(temp.value, 2);
    requestData += ",\"humidity\":";
    requestData += String(temp.humidity, 2);
    requestData += "}";
    int status = http.POST(requestData);
    if (status > 0) {
        String payload = http.getString();
        if (status >= 400) {
            String title;
            title.reserve(9);
            title += "HTTP ";
            title += String(status);
            display.setError({title, payload});
            display.render();
            retryRequest = true;
        } else {
            display.clearError();
            Serial.printf("HTTP Response code: %d\nBody: %s\n", status,
                          payload.c_str());
            retryRequest = false;
        }
    } else {
        display.setError({.title = "HTTP req failed",
                          .message = http.errorToString(status)});
        display.render();
        retryRequest = true;
    }
    http.end();
}
