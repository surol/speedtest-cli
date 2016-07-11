package speedtest

import (
	"log"
	"errors"
	"sort"
	"fmt"
)

type Server struct {
	Coordinates
	URL      string `xml:"url,attr"`
	Name     string `xml:"name,attr"`
	Country  string `xml:"country,attr"`
	CC       string `xml:"cc,attr"`
	Sponsor  string `xml:"sponsor,attr"`
	ID       uint64 `xml:"id,attr"`
	URL2     string `xml:"url2,attr"`
	Host     string `xml:"host,attr"`
	Distance float64 `xml:"-"`
}

func (s *Server) String() string {
	return fmt.Sprintf("%d %s (%s, %s) --- %.2f km", s.ID, s.Sponsor, s.Name, s.Country, s.Distance)
}

type Servers struct {
	List []*Server `xml:"servers>server"`
}

func (ss *Servers) Len() int {
	return len(ss.List)
}

func (ss *Servers) Less(i, j int) bool {
	return ss.List[i].Distance < ss.List[j].Distance
}

func (ss *Servers) Swap(i, j int) {
	temp := ss.List[i]
	ss.List[i] = ss.List[j]
	ss.List[j] = temp;
}

func (ss *Servers) Truncate(len uint) {
	ss.List = ss.List[:len]
}

func (ss *Servers) String() string {
	out := ""
	for _, server := range ss.List {
		out += server.String() + "\n"
	}
	return out
}

var serverURLs = [...]string{
	"://www.speedtest.net/speedtest-servers-static.php",
	"://c.speedtest.net/speedtest-servers-static.php",
	"://www.speedtest.net/speedtest-servers.php",
	"://c.speedtest.net/speedtest-servers.php",
}

var NoServersError error = errors.New("No servers available")

func (client *Client) Servers() (servers *Servers, err error) {
	if client.servers != nil {
		return client.servers, nil
	}

	config, err := client.Config();
	if err != nil {
		return nil, err
	}

	client.Log("Retrieving speedtest.net server list...")

	servers = &Servers{}
	for _, url := range serverURLs {
		resp, err := client.Get(url)
		if resp != nil {
			url = resp.Request.URL.String()
		}
		if err != nil {
			log.Printf("Failed to retrieve server list from %s: %v", url, err)
		}


		if err = resp.ReadXML(servers); err != nil {
			log.Printf("Failed to read server list %s: %v", url, err)
		}
	}

	if servers == nil {
		return nil, NoServersError
	}
	for _, server := range servers.List {
		server.Distance = server.DistanceTo(config.Client.Coordinates)
	}
	sort.Sort(servers)
	client.servers = servers

	return servers, nil;
}
