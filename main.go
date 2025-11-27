package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mucusscraper/pokedex_go/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Config struct {
	Actual           string
	Pokemon          string
	PokemonToInspect []string
	Next             *string
	Previous         *string
}

type LocationAreaEndPointResults struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type LocationAreaEndPoint struct {
	Count    int                           `json:"count"`
	Next     *string                       `json:"next"`
	Previous *string                       `json:"previous"`
	Results  []LocationAreaEndPointResults `json:"results"`
}

type PokeRef struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}
type PokemonEncounter struct {
	Pokemon PokeRef `json:"pokemon"`
}
type ExploreResponse struct {
	Pokemons []PokemonEncounter `json:"pokemon_encounters"`
}

type PokemonStats struct {
	Stats    PokeRef `json:"stat"`
	BaseStat int     `json:"base_stat"`
}

type PokemonTypes struct {
	Type PokeRef `json:"type"`
}

type Pokemon struct {
	Name           string         `json:"name"`
	BaseExperience int            `json:"base_experience"`
	Height         int            `json:"height"`
	Weight         int            `json:"weight"`
	Stats          []PokemonStats `json:"stats"`
	Types          []PokemonTypes `json:"types"`
}

var map_of_commands map[string]cliCommand
var map_of_pokemons_caught map[string]Pokemon
var cache = pokecache.NewCache(10 * time.Second)

func main() {
	config := &Config{}
	map_of_pokemons_caught = map[string]Pokemon{}
	map_of_commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Maps the 20 next locations",
			callback:    commandMap,
		},
		"bmap": {
			name:        "bmap",
			description: "Maps the 20 previous locations",
			callback:    commandBMap,
		},
		"explore": {
			name:        "explore",
			description: "Lists the pokemons found in a certain area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Tries to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect your pokemons",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Shows your pokedex",
			callback:    commandPokedex,
		},
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Pokedex > ")
		if scanner.Scan() {
			all_string := scanner.Text()
			result_from_clean_input := cleanInput(all_string)
			if result_from_clean_input[0] != "explore" && result_from_clean_input[0] != "catch" && result_from_clean_input[0] != "inspect" {
				first_value := result_from_clean_input[0]
				_, ok := map_of_commands[first_value]
				if !ok {
					fmt.Printf("Unknown command\n")
				} else {
					map_of_commands[first_value].callback(config)
				}
			} else {
				_, ok := map_of_commands[result_from_clean_input[0]]
				if !ok {
					fmt.Printf("Unknown command\n")
				} else {
					if result_from_clean_input[0] == "explore" {
						config.Actual = result_from_clean_input[1]
						map_of_commands[result_from_clean_input[0]].callback(config)
					}
					if result_from_clean_input[0] == "catch" {
						config.Pokemon = result_from_clean_input[1]
						map_of_commands[result_from_clean_input[0]].callback(config)
					}
					if result_from_clean_input[0] == "inspect" {
						config.PokemonToInspect = append(config.PokemonToInspect, result_from_clean_input[1])
						map_of_commands[result_from_clean_input[0]].callback(config)
					}
				}
			}
		}
	}
}

func cleanInput(text string) []string {
	elements := strings.Fields(strings.ToLower(text))
	if elements[0] == "explore" {
		new_list := []string{elements[0], elements[1]}
		return new_list
	} else {
		return elements
	}
}

func commandExit(config *Config) error {
	fmt.Printf("Closing the Pokedex... Goodbye!\n")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\n")
	for key, value := range map_of_commands {
		fmt.Printf("%s: %s\n", key, value.description)
	}
	return nil
}
func getIssues(url string) (*LocationAreaEndPoint, error) {
	if data, ok := cache.Get(url); ok {
		var issue *LocationAreaEndPoint
		if err := json.Unmarshal(data, &issue); err != nil {
			return nil, err
		}
		return issue, nil
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var issue *LocationAreaEndPoint
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, err
	}
	cache.Add(url, data)
	return issue, nil

}
func commandMap(config *Config) error {
	if config.Next == nil {
		res, _ := getIssues("https://pokeapi.co/api/v2/location-area/")
		for _, location := range res.Results {
			fmt.Printf("%v\n", location.Name)
		}
		config.Next = res.Next
		config.Previous = res.Previous
	} else {
		res, _ := getIssues(*config.Next)
		for _, location := range res.Results {
			fmt.Printf("%v\n", location.Name)
		}
		config.Next = res.Next
		config.Previous = res.Previous
	}
	return nil
}

func commandBMap(config *Config) error {
	if config.Previous == nil {
		fmt.Print("Not previous maps\n")
	} else {
		res, _ := getIssues(*config.Previous)
		for _, location := range res.Results {
			fmt.Printf("%v\n", location.Name)
		}
		config.Next = res.Next
		config.Previous = res.Previous
	}
	return nil
}

func getIssuesPokeEncounters(url string) (*ExploreResponse, error) {
	if data, ok := cache.Get(url); ok {
		var issue *ExploreResponse
		if err := json.Unmarshal(data, &issue); err != nil {
			return nil, err
		}
		return issue, nil
	} else {
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var issue *ExploreResponse
		if err := json.Unmarshal(data, &issue); err != nil {
			return nil, err
		}
		cache.Add(url, data)
		return issue, nil
	}
}

func commandExplore(config *Config) error {
	if config.Actual == "" {
		fmt.Printf("No area to explore pokemons!\n")
	} else {
		full_url := "https://pokeapi.co/api/v2/location-area/" + config.Actual
		res, _ := getIssuesPokeEncounters(full_url)
		fmt.Printf("Exploring %s\n", config.Actual)
		fmt.Printf("Found Pokemon:\n")
		for _, pokemon := range res.Pokemons {
			fmt.Printf("- %v\n", pokemon.Pokemon.Name)
		}
		config.Next = nil
	}
	return nil
}

func getIssuesPokemon(url string) (*Pokemon, error) {
	if data, ok := cache.Get(url); ok {
		var issue *Pokemon
		if err := json.Unmarshal(data, &issue); err != nil {
			return nil, err
		}
		return issue, nil
	} else {
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		var issue *Pokemon
		if err := json.Unmarshal(data, &issue); err != nil {
			return nil, err
		}
		cache.Add(url, data)
		return issue, nil
	}
}

func commandCatch(config *Config) error {
	if config.Pokemon == "" {
		fmt.Printf("No pokemon to catch!\n")
	} else {
		full_url := "https://pokeapi.co/api/v2/pokemon/" + config.Pokemon
		res, _ := getIssuesPokemon(full_url)
		poke_name := res.Name
		fmt.Printf("Throwing a Pokeball at %s...\n", poke_name)
		random_number := rand.Float64()
		chance_to_catch := 1.0 - (float64(res.BaseExperience) / 650.0)
		if chance_to_catch < random_number {
			fmt.Printf("%s escaped!\n", poke_name)
		} else {
			fmt.Printf("%s was caught!\n", poke_name)
			map_of_pokemons_caught[poke_name] = *res
		}
	}
	return nil
}

func commandInspect(config *Config) error {
	for key, value := range map_of_pokemons_caught {
		if key == config.PokemonToInspect[len(config.PokemonToInspect)-1] {
			fmt.Printf("Name: %s\n", key)
			fmt.Printf("Height: %d\n", value.Height)
			fmt.Printf("Weight: %d\n", value.Weight)
			fmt.Printf("Stats: \n")
			for _, value_stats := range value.Stats {
				fmt.Printf("  -%s: %d\n", value_stats.Stats.Name, value_stats.BaseStat)
			}
			fmt.Printf("Types: \n")
			for _, value_types := range value.Types {
				fmt.Printf("  -%s\n", value_types.Type.Name)
			}
			return nil
		}
	}
	fmt.Printf("Pokemon not caught!\n")
	return nil
}

func commandPokedex(config *Config) error {
	fmt.Printf("Your Pokedex:\n")
	for key, _ := range map_of_pokemons_caught {
		fmt.Printf(" - %s\n", key)
	}
	return nil
}
