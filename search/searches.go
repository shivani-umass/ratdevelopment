package searching

import (
	"regexp"
	"strconv"
	"strings"
)

//SearchContext is the struct to be used for searching
type SearchContext struct {
	Systems []System
}

//System contains all the information that can be searched on in a search query
type System struct {
	id       string
	capacity int
	tenants  []string
}

//takes a system in and narrows by a string value of a property
type proposure func([]System) []System

//Search takes a search query and returns a narrowed list of systems by that query
func (context *SearchContext) Search(query string) []System {
	proposures := enclose(query)
	results := make([]System, len(context.Systems))
	copy(results, context.Systems)
	for _, prop := range proposures {
		results = prop(results)
	}
	return results
}

//enclose takes the query and splits it into singular queries
func enclose(query string) []proposure {
	query = strings.Replace(query, " ", "", -1)
	narrows := strings.Split(query, ",")
	props := make([]proposure, 0)
	for _, narrow := range narrows {
		pNarrow := regexp.MustCompile("[<>=]").Split(narrow, 2)
		if len(pNarrow) == 1 { //singular search, like "011890" or "hpe"
			first := narrow[0]
			if '0' <= first && first <= '9' { //is a system ID
				props = append(props, narrowID(narrow))
			} else { //is a tenant
				props = append(props, narrowTenant(narrow))
			}
		} else {
			//capacity := pNarrow[1]
			comp := narrow[len(pNarrow[0])]
			//fmt.Println(compVal)
			iCap, _ := strconv.Atoi(pNarrow[1])
			props = append(props, narrowCapacity(iCap, comp))
		}
	}
	return props
}

func narrowID(id string) proposure {
	return func(systems []System) []System {
		results := make([]System, 0)
		for _, system := range systems {
			if strings.Contains(system.id, id) {
				results = append(results, system)
			}
		}
		return results
	}
}

//tenant may just be a substring of a tenant
func narrowTenant(tenant string) proposure {
	return func(systems []System) []System {
		results := make([]System, 0)
		for _, system := range systems {
			for _, ten := range system.tenants {
				if strings.Contains(ten, tenant) {
					results = append(results, system)
					break
				}
			}
		}
		return results
	}
}

func narrowCapacity(capacity int, comp byte) proposure {
	return func(systems []System) []System {
		results := make([]System, 0)
		for _, system := range systems {
			if (comp == '=' && system.capacity == capacity) || (comp == '<' && system.capacity < capacity) || (comp == '>' && system.capacity > capacity) {
				results = append(results, system)
			}
		}
		return results
	}
}

//compares the signs of two integers.  If both are positive or both are negative, or
func compSign(a, compVal int) bool {
	if compVal == 0 { //if compVal is 0, then the two values should be equal, so (expected - actual) == 0
		return a == 0
	}
	return a*compVal >= 0 //otherwise, do a comparison
}
