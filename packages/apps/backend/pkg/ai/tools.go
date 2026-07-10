package ai

type ToolDefinition[Input any, Output any] struct {
	Name string
}

type (
	WeatherToolInput struct {
		Location string `json:"location"`
	}
	WeatherToolOutput struct {
		DegreesCelcius string `json:"degrees_celcius"`
	}
	WeatherToolDefinition ToolDefinition[WeatherToolInput, WeatherToolOutput]
)
