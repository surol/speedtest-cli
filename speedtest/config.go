package speedtest

import (
	"io/ioutil"
	"encoding/xml"
	"strings"
	"strconv"
	"log"
)

type ClientConfig struct {
	IP                 string `xml:"ip,attr"`
	Latitude           float32 `xml:"lat"`
	Longitude          float32 `xml:"lon"`
	ISP                string `xml:"isp"`
	ISPRating          float32 `xml:"isprating"`
	ISPDownloadAverage uint32 `xml:"ispdlavg"`
	ISPUploadAverage   uint32 `xml:"ispulavg"`
	Rating             float32 `xml:"rating"`
	LoggedIn           uint8 `xml:"loggedin"`
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

	xmlcontent, err := ioutil.ReadAll(resp.Body)
	cerr := resp.Body.Close();
	if err != nil {
		return nil, err
	}
	if cerr != nil {
		return nil, cerr
	}

	config = &Config{}

	err = xml.Unmarshal(xmlcontent, config)
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

