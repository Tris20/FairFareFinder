Using gpt and our formula and modifier tables to estimate prices for the following flights, 2 days from today, 


price from formula| formula price with LFM | Route| cheapest flight(one way)| number of flights between 29/1/25 and 2/2/25 | airline | 
41 | 48  | ber - cph | 45  | 12 | 
  | ber - cgn | 100 | 17 |
  | ber - dus | 90  | 12 |
 50 | 58 | ber - muc | 130 | 39 |
---------------
  | gla - bhx | 40  | 5  |
  | gla - cal | 83  | 6  | 
 84 | 98 | gla - lhr | 110 | 43 | BA
  | gla - syy | 250 | 9  |
  | gla - bfs | 40  | 10 | easyjet
---------------
 101 | 123 | edi - SYY | 240 | 3  | loganair
  | edi - lhr | 120 | 44 |
---------------
 182 | 207 |  bre - ist | 120 | 6  | turkish air 




sql query to get number of flights in timerange 
SELECT *
FROM schedule
WHERE departureAirport = 'BER'
AND arrivalAirport = 'CPH'
AND DATE(departureTime) BETWEEN '2025-01-29' AND '2025-02-02'
ORDER BY departureTime;



time of day modifier? Need to classify each route as either tourist or busiuness: tourist route is cheaper at 6am, business route is more expensive at 6am


day of week modifier
