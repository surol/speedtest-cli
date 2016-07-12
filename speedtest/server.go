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

type ServersRef struct {
	Servers *Servers
	Error error
}

func (servers *Servers) Len() int {
	return len(servers.List)
}

func (servers *Servers) Less(i, j int) bool {
	server1 := servers.List[i]
	server2 := servers.List[j]
	if server1.ID == server2.ID {
		return false;
	}
	if server1.Distance < server2.Distance {
		return true;
	}
	if server1.Distance > server2.Distance {
		return false
	}
	return server1.ID < server2.ID
}

func (servers *Servers) Swap(i, j int) {
	temp := servers.List[i]
	servers.List[i] = servers.List[j]
	servers.List[j] = temp;
}

func (servers *Servers) Truncate(max int) {
	size := servers.Len()
	if size < max {
		max = size;
	}
	servers.List = servers.List[:max]
}

func (servers *Servers) String() string {
	out := ""
	for _, server := range servers.List {
		out += server.String() + "\n"
	}
	return out
}

func (servers *Servers) append(other *Servers) *Servers {
	if servers == nil {
		return other
	}
	servers.List = append(servers.List, other.List...)
	return servers
}

func (servers *Servers) sort(config *Config) {
	for _, server := range servers.List {
		server.Distance = server.DistanceTo(config.Client.Coordinates)
	}
	sort.Sort(servers)
}

func (servers *Servers) deduplicate() {
	dedup := make([]*Server, 0, len(servers.List));
	var prevId  uint64 = 0;
	for _, server := range servers.List {
		if prevId != server.ID {
			prevId = server.ID
			dedup = append(dedup, server);
		}
	}
	servers.List = dedup
}

var serverURLs = [...]string{
	"://www.speedtest.net/speedtest-servers-static.php",
	"://c.speedtest.net/speedtest-servers-static.php",
	"://www.speedtest.net/speedtest-servers.php",
	"://c.speedtest.net/speedtest-servers.php",
}

var NoServersError error = errors.New("No servers available")

func (client *Client) AllServers(ret chan ServersRef) {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if client.allServers == nil {
		client.allServers = make(chan ServersRef)
		go client.loadServers()
	}

	go func() {
		result := <- client.allServers
		ret <- result
		client.allServers <- result// Make it available again
	}()
}

func (client *Client) loadServers() {
	configChan := make(chan ConfigRef)
	client.Config(configChan);

	client.Log("Retrieving speedtest.net server list...")

	serversChan := make(chan *Servers, len(serverURLs))
	for _, url := range serverURLs {
		go client.loadServersFrom(url, serversChan)
	}

	var servers *Servers

	for range serverURLs {
		servers = servers.append(<- serversChan);
	}

	result := ServersRef{}

	if servers.Len() == 0 {
		result.Error = NoServersError
	} else {
		configRef := <- configChan
		if configRef.Error != nil {
			result.Error = configRef.Error
		} else {
			servers.sort(configRef.Config)
			servers.deduplicate()
			result.Servers = servers
		}
	}

	client.allServers <- result
}

func (client *Client) loadServersFrom(url string, ret chan *Servers) {
	resp, err := client.Get(url)
	if resp != nil {
		url = resp.Request.URL.String()
	}
	if err != nil {
		log.Printf("Failed to retrieve server list from %s: %v", url, err)
	}

	servers := &Servers{}
	if err = resp.ReadXML(servers); err != nil {
		log.Printf("Failed to read server list %s: %v", url, err)
	}
	ret <- servers
}
