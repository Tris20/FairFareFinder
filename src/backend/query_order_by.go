package backend

var orderByClauses = map[string]string{
	"cheapest_fnaf":         "ORDER BY fnf.price_fnaf ASC",
	"most_expensive_fnaf":   "ORDER BY fnf.price_fnaf DESC",
	"best_weather":          "ORDER BY avg_wpi DESC",
	"worst_weather":         "ORDER BY avg_wpi ASC",
	"cheapest_hotel":        "ORDER BY a.booking_pppn ASC",
	"most_expensive_hotel":  "ORDER BY a.booking_pppn DESC",
	"shortest_flight":       "ORDER BY f.duration_hour_dot_mins ASC",
	"longest_flight":        "ORDER BY f.duration_hour_dot_mins DESC",
	"cheapest_flight":       "ORDER BY f.price_this_week ASC",
	"most_expensive_flight": "ORDER BY f.price_this_week DESC",
}

func determineOrderClause(sortOption string) string {
	if clause, found := orderByClauses[sortOption]; found {
		return clause
	}
	return "ORDER BY avg_wpi DESC" // Default
}
