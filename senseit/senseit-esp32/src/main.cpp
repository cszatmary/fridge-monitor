#include "secrets.h"
#include <Adafruit_BME280.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>
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
TwoWire screenWire = TwoWire(0);
TwoWire sensorWire = TwoWire(1);
Adafruit_SSD1306 screen =
    Adafruit_SSD1306(SCREEN_WIDTH, SCREEN_HEIGHT, &screenWire);
Adafruit_BME280 sensor;

// scani2c is a small helper function to find all available i2c devices and
// their address on the i2c bus used by wire.
void scani2c(TwoWire &wire) {
    Serial.println("Scanning...");
    int numDevices = 0;
    for (byte address = 1; address < 127; address++) {
        wire.beginTransmission(address);
        byte error = wire.endTransmission();
        if (error == 0) {
            Serial.print("I2C device found at address 0x");
            if (address < 16) {
                Serial.print("0");
            }
            Serial.println(address, 16);
            numDevices++;
        } else if (error == 4) {
            Serial.print("Unknow error at address 0x");
            if (address < 16) {
                Serial.print("0");
            }
            Serial.println(address, 16);
        }
    }
    if (numDevices == 0) {
        Serial.println("No I2C devices found\n");
        return;
    }
    Serial.println("Done scanning for I2C devices\n");
}

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

    // Setup WiFi on boot.
    // Register event handlers so that it automatically reconnects if the
    // connection is lost.
    WiFi.onEvent(handleWiFiStationConnected,
                 WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_CONNECTED);
    WiFi.onEvent(handleWiFiGotIP, WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_GOT_IP);
    WiFi.onEvent(handleWiFiStationDisconnected,
                 WiFiEvent_t::ARDUINO_EVENT_WIFI_STA_DISCONNECTED);
    WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.println("Connecting to WiFi...");
    }

    // Setup I2C devices.
    screenWire.begin(SCREEN_I2C_SDA, SCREEN_I2C_SCL);
    sensorWire.begin(SENSOR_I2C_SDA, SENSOR_I2C_SCL, 100000U);
    if (!screen.begin(SSD1306_SWITCHCAPVCC, SCREEN_I2C_ADDRESS)) {
        Serial.println("SSD1306 initialization failed");
        // Nothing we can really do here.
        while (true) {
            delay(10000);
        }
    }
    screen.clearDisplay();
    if (!sensor.begin(SENSOR_I2C_ADDRESS, &sensorWire)) {
        // Since the screen was initialized we can use it to display an error.
        screen.setTextColor(WHITE);
        screen.setTextSize(1);
        screen.setCursor(0, 0);
        screen.print("Error: BME280 sensor initialization failed");
        screen.display();
        Serial.println("BME280 sensor initialization failed");
        while (true) {
            delay(10000);
        }
    }
}

void loop() {
    if ((millis() - lastTimeUpdate) < UPDATE_INTERVAL_MS) {
        return;
    }
    lastTimeUpdate = millis();

    // Uncomment these if you need to locate the i2c addresses.
    // scani2c(screenWire);
    // scani2c(sensorWire);

    // Always get the temperature and humidity and display it on the screen.
    float temp = sensor.readTemperature();
    float humidity = sensor.readHumidity();
    Serial.printf("Temperature = %.2f\nHumidity %.2f%\n", temp, humidity);
    screen.clearDisplay();
    screen.setTextColor(WHITE);
    // Display temperature
    screen.setTextSize(1);
    screen.setCursor(0, 0);
    screen.print("Temperature: ");
    screen.setTextSize(2);
    screen.setCursor(0, 10);
    screen.print(temp);
    screen.print(" ");
    screen.setTextSize(1);
    screen.cp437(true);
    screen.write(167); // Degrees symbol
    screen.setTextSize(2);
    screen.print("C");
    // Display humidity
    screen.setTextSize(1);
    screen.setCursor(0, 35);
    screen.print("Humidity: ");
    screen.setTextSize(2);
    screen.setCursor(0, 45);
    screen.print(humidity);
    screen.print(" %");
    screen.display();

    if ((millis() - lastTimePost) < POST_INTERVAL_MS) {
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
    String requestData = "{\"value\":" + String(temp, 2) +
                         ",\"humidity\":" + String(humidity, 2) + "}";
    int status = http.POST(requestData);
    if (status > 0) {
        String payload = http.getString();
        Serial.printf("HTTP Response code: %d\nBody: %s\n", status,
                      payload.c_str());
    } else {
        Serial.printf("HTTP error: %s\n", http.errorToString(status).c_str());
    }
    http.end();
}
