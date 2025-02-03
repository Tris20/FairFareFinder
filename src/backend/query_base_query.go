package backend

const BaseQuery = `
    SELECT 
        ds.destination_city_name,
        MIN(f.price_next_week) AS price_city1,
        MIN(f.skyscanner_url_next_week) AS url_city1,
        w.date,
        w.avg_daytime_temp,
        w.weather_icon,
        w.google_url,
        l.avg_wpi,
        l.image_1,
        a.booking_url,
        a.booking_pppn,
        fnf.price_fnaf,
        MIN(f.duration_in_minutes) AS duration_mins,
        MIN(f.duration_in_hours) AS duration_hours,
        MIN(f.duration_in_hours_rounded) AS duration_hours_rounded,
        MIN(f.duration_hour_dot_mins) AS duration_hour_dot_mins
    FROM DestinationSet ds
    JOIN flight f ON ds.destination_city_name = f.destination_city_name 
                   AND ds.destination_country = f.destination_country
    JOIN location l ON ds.destination_city_name = l.city 
                     AND ds.destination_country = l.country
    JOIN weather w ON w.city = ds.destination_city_name 
                    AND w.country = ds.destination_country
    LEFT JOIN accommodation a ON a.city = ds.destination_city_name 
                               AND a.country = ds.destination_country
    LEFT JOIN (
        SELECT 
            fnf.origin_city,
            fnf.origin_country,
            fnf.destination_city,
            fnf.destination_country,
            MIN(fnf.price_fnaf) AS price_fnaf
        FROM five_nights_and_flights fnf
        GROUP BY fnf.origin_city, fnf.origin_country, fnf.destination_city, fnf.destination_country
    ) fnf ON fnf.destination_city = ds.destination_city_name
           AND fnf.destination_country = ds.destination_country
           AND fnf.origin_city = f.origin_city_name
           AND fnf.origin_country = f.origin_country
    WHERE l.avg_wpi BETWEEN 1.0 AND 10.0 
      AND w.date >= date('now')
      AND f.price_next_week < ?
      AND f.origin_city_name IN
`
