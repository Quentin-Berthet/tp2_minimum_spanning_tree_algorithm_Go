package neighbour

import (
	"TP2_Minimum_Spanning_Tree/environment"
	"TP2_Minimum_Spanning_Tree/logger"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Type string

const (
	Branch   Type = "Branch"
	Rejected Type = "Rejected"
	Basic    Type = "Basic"
)

type Neighbour struct {
	Ip     string
	Port   string
	Weight int
	Type   Type
}

func New(ip string, port string, weight int) *Neighbour {
	return &Neighbour{
		Ip:     ip,
		Port:   port,
		Weight: weight,
		Type:   Basic,
	}
}

func parseFileLine(line string) *Neighbour {
	line = strings.TrimSpace(line)
	if line[0] == '#' {
		return nil
	}
	splitLine := strings.Split(line, " ")
	weight, err := strconv.Atoi(splitLine[2])
	if err != nil {
		logger.Log(parseFileLine, fmt.Sprintf("'%s' n'est pas un numéro de port valide.", splitLine[2]))
		return nil
	}
	return New(
		splitLine[0],
		splitLine[1],
		weight,
	)
}

func LoadFromFile(filename string) []*Neighbour {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Log(LoadFromFile, fmt.Sprintf("%s", err.Error()))
		os.Exit(1)
	}
	content := strings.Trim(string(data), "\n")
	lines := strings.Split(content, "\n")
	neighbours := make([]*Neighbour, 0)
	for _, line := range lines {
		neighbour := parseFileLine(line)
		if neighbour != nil {
			neighbours = append(neighbours, neighbour)
		}
	}
	return neighbours
}

func SortNeighbours(neighbours []*Neighbour) {
	sort.SliceStable(neighbours, func(i, j int) bool {
		return neighbours[i].Weight < neighbours[j].Weight
	})
}

func NeighboursToMap(neighbours []*Neighbour, env environment.Environment) map[string]*Neighbour {
	neighboursMap := make(map[string]*Neighbour, 0)
	for _, neighbour := range neighbours {
		if env == environment.Cloud {
			neighboursMap[neighbour.Ip] = neighbour
		} else {
			neighboursMap[neighbour.Port] = neighbour
		}
	}
	return neighboursMap
}

func (neighbour Neighbour) ToAddressPort() string {
	return fmt.Sprintf("%s:%s", neighbour.Ip, neighbour.Port)
}

func (neighbour Neighbour) ToString() string {
	return fmt.Sprintf("Neighbour(%+v)", neighbour)
}
