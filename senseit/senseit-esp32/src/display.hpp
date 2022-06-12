#ifndef DISPLAY_H
#define DISPLAY_H

#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>
#include <Wire.h>

struct Temperature {
    float value;
    float humidity;
};

struct Content {
    String title;
    String message;
};

class Display {
  private:
    Adafruit_SSD1306 screen;

    // Render specific state
    Temperature temp;
    bool tempSet;
    Content info;
    bool infoSet;
    Content error;
    bool errorSet;
    uint8_t currLine;
    uint8_t textSize;

    void nextLine();

  public:
    float minSafeTemp;
    float maxSafeTemp;
    bool printToSerial;

    bool begin(TwoWire *wire, uint8_t i2caddr, uint8_t w, uint8_t h);
    void setTemperature(Temperature t);
    void setInfo(Content c);
    void setError(Content c);
    void clear();
    void clearError();
    void render();
};

#endif // DISPLAY_H
