#include "display.hpp"

float celsiusToFahrenheit(float c) { return (c * 1.8) + 32; }

bool Display::begin(TwoWire *wire, uint8_t i2caddr, uint8_t w, uint8_t h) {
    this->screen = Adafruit_SSD1306(w, h, wire);
    if (!this->screen.begin(SSD1306_SWITCHCAPVCC, i2caddr)) {
        return false;
    }
    this->clear();
    // Set super large/small defaults to ensure always OK status if not set.
    this->minSafeTemp = -10000;
    this->maxSafeTemp = 10000;
    this->printToSerial = false;
    return true;
}

void Display::setTemperature(Temperature t) {
    this->temp = t;
    this->tempSet = true;
}

void Display::setInfo(Content c) {
    this->info = c;
    this->infoSet = true;
}

void Display::setError(Content c) {
    this->error = c;
    this->errorSet = true;
}

void Display::clear() {
    this->tempSet = false;
    this->infoSet = false;
    this->errorSet = false;
    this->screen.clearDisplay();
}

void Display::clearError() { this->errorSet = false; }

void Display::render() {
    // Make sure we clear the previous contents before proceeding.
    this->screen.clearDisplay();
    // If no content to display then no-op (i.e. display nothign).
    if (!this->tempSet && !this->infoSet && !this->errorSet) {
        return;
    }
    this->screen.setTextColor(WHITE);
    this->screen.setTextWrap(false);
    this->currLine = 0;
    this->textSize = 1;

    // Handle just temperature.
    if (this->tempSet && !this->infoSet && !this->errorSet) {
        // Use a bigger text size so it looks nicer and is easier to read.
        this->textSize = 2;
        this->screen.setTextSize(textSize);
        // First line is the status.
        this->nextLine();
        if (this->temp.value < this->minSafeTemp) {
            this->screen.print(" TOO COLD");
        } else if (this->temp.value > this->maxSafeTemp) {
            this->screen.print("  TOO HOT");
        } else {
            this->screen.print("    OK");
        }
        // Second line is the temperature in both C and F.
        this->nextLine();
        this->screen.print(this->temp.value, 1);
        this->screen.print("  ");
        this->screen.print(celsiusToFahrenheit(this->temp.value), 1);
        // Third line is the humidity.
        this->nextLine();
        this->screen.print("    ");
        this->screen.print(this->temp.humidity, 0);
        this->screen.print("%");
        this->screen.display();
        if (this->printToSerial) {
            Serial.printf("Temperature: %.1fC, humidity: %.0f%\n",
                          this->temp.value, this->temp.humidity);
        }
        return;
    }

    this->screen.setTextSize(textSize);
    if (this->infoSet) {
        this->nextLine();
        this->screen.print(this->info.title);
        this->nextLine();
        this->screen.print(this->info.message);
        if (this->printToSerial) {
            Serial.printf("Info: %s\n%s\n", this->info.title.c_str(),
                          this->info.message.c_str());
        }
    }
    if (this->tempSet) {
        this->nextLine();
        if (this->temp.value < this->minSafeTemp) {
            this->screen.print("    TOO COLD");
        } else if (this->temp.value > this->maxSafeTemp) {
            this->screen.print("        TOO HOT");
        } else {
            this->screen.print("                OK");
        }
        this->nextLine();
        // Display both temperature and humidity on the same line so we have
        // more screen space for a potential error.
        this->screen.print(this->temp.value, 1);
        this->screen.print("    ");
        this->screen.print(celsiusToFahrenheit(this->temp.value), 1);
        this->screen.print("    ");
        this->screen.print(this->temp.humidity, 0);
        this->screen.print('%');
        if (this->printToSerial) {
            Serial.printf("Temperature: %.1fC, humidity: %.0f%\n",
                          this->temp.value, this->temp.humidity);
        }
    }
    if (this->errorSet) {
        this->nextLine();
        this->screen.print("Error: ");
        this->screen.print(this->error.title);
        this->nextLine();
        this->screen.print(this->error.message);
        if (this->printToSerial) {
            Serial.printf("Error: %s\n%s\n", this->error.title.c_str(),
                          this->error.message.c_str());
        }
    }
    this->screen.display();
}

void Display::nextLine() {
    this->screen.setCursor(0, this->currLine * this->textSize * 8);
    this->currLine++;
}
