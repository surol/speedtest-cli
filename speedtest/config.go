package speedtest

import (
	"encoding/xml"
	"strings"
	"strconv"
	"log"
)

type ClientConfig struct {
	Coordinates
	IP                 string `xml:"ip,attr"`
	ISP                string `xml:"isp,attr"`
	ISPRating          float32 `xml:"isprating,attr"`
	ISPDownloadAverage uint32 `xml:"ispdlavg,attr"`
	ISPUploadAverage   uint32 `xml:"ispulavg,attr"`
	Rating             float32 `xml:"rating,attr"`
	LoggedIn           uint8 `xml:"loggedin,attr"`
}

type ConfigTime struct {
	Upload   uint32
	Download uint32
}

type ConfigTimes []ConfigTime

type Config struct {
	Client ClientConfig `xml:"client"`
	Times  ConfigTimes `xml:"times"`
}

func (client *Client) Log(format string, a ...interface{}) {
	if !client.opts.Simple {
		log.Printf(format, a...)
	}
}

func (client *Client) Config() (config *Config, err error) {
	if client.config != nil {
		return client.config, nil
	}

	client.Log("Retrieving speedtest.net configuration...")

	resp, err := client.Get("://www.speedtest.net/speedtest-config.php")
	if err != nil {
		return nil, err
	}

	config = &Config{}
	err = resp.ReadXML(config)
	if err != nil {
		return nil, err
	}

	client.config = config

	return config, nil
}

func (times ConfigTimes) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, attr := range start.Attr {
		name := attr.Name.Local
		if dl := strings.HasPrefix(name, "dl"); dl || strings.HasPrefix(name, "ul") {
			num, err := strconv.Atoi(name[2:])
			if err != nil {
				return err;
			}
			if num > cap(times) {
				newTimes := make([]ConfigTime, num)
				copy(newTimes, times)
				times = newTimes[0:num]
			}

			speed, err := strconv.ParseUint(attr.Value, 10, 32);

			if err != nil {
				return err
			}
			if dl {
				times[num - 1].Download = uint32(speed)
			} else {
				times[num - 1].Upload = uint32(speed)
			}
		}
	}

	return d.Skip()
}

