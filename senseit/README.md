# senseit

`senseit` is a small application that reads the temperature from a fridge and sends it in a POST request to a destination of your choosing.
It was built to run on a [Raspberry Pi Zero W](https://www.raspberrypi.com/products/raspberry-pi-zero-w/) and uses a
[BME280](https://www.adafruit.com/product/2652) sensor to read the current temperature.

## Usage

First build it by running the following command. This will produce a binary called `senseit`.

```sh
GOOS=linux GOARCH=arm go build -ldflags "-s -w"
```

I found it easiest to develop and build on my mac thanks to go's cross-compile capabilities. Then I copied it to my raspberry pi using `scp`.

To use it, I set up a cron job using `crontab` that runs `senseit` every 10min to send a temperature to my monitorit server.

Here's an example configuration:

```
*/10 * * * * /home/pi/senseit 0x76 <URL>/fridges/1/temperatures >> /home/pi/senseit.log 2>&1
```
