**Final Flight Price Formula:**

\[
\text{PRICE}_{\text{final}} = (\text{BA} \times \text{FD}) \times \text{AM} \times \text{PM} \times \text{DM} \times \text{FFM} \times \text{SNM} \times \text{ACM} \times \text{RCM}
\]

**Where:**

- **BA (Base Price per Minute):**  
  The fundamental cost per minute of flight (e.g., €0.15/min).

- **FD (Flight Duration):**  
  The total duration of the flight in minutes.

- **AM (Airline Multiplier):**  
  A factor reflecting the airline's relative pricing level (e.g., low-cost carriers might have values below 1, while full-service airlines might be above 1).

- **PM (Population Modifier):**  
  A multiplier based on the populations of the origin and destination cities (derived from a population bracket table).

- **DM (Date/Season/Holiday Modifier):**  
  A factor that increases prices during peak travel periods (such as holidays or school breaks).

- **FFM (Flight-Frequency Modifier):**  
  A modifier based on the number of direct flights available over a set time period (fewer flights usually mean higher prices).

- **SNM (Short-Notice Modifier):**  
  A factor reflecting how many days before departure the booking is made (last-minute bookings typically have higher multipliers).

- **ACM (Aircraft Capacity Multiplier):**  
  A multiplier based on the size or seating capacity of the aircraft (smaller aircraft tend to have higher per-seat costs).

- **RCM (Route Classification Multiplier):**  
  A factor that adjusts for the nature of the route (e.g., pure business, mixed business/leisure, or pure leisure) to capture qualitative differences in demand.




# Multiplier Tables for Flight Price Estimation

---

## 1. Airline Multipliers

| Airline                        | Price Multiplier |
|--------------------------------|------------------|
| Ryanair                        | 0.6              |
| Wizz Air                       | 0.7              |
| EasyJet                        | 0.8              |
| Allegiant Air                  | 0.8              |
| Frontier Airlines              | 0.8              |
| Spirit Airlines                | 0.8              |
| Southwest Airlines             | 0.9              |
| JetBlue                        | 1.0              |
| Alaska Airlines                | 1.0              |
| AirAsia                        | 1.0              |
| IndiGo                         | 1.0              |
| Vueling                        | 1.0              |
| Norwegian                      | 1.0              |
| Scoot                          | 1.0              |
| Delta Air Lines                | 1.2              |
| United Airlines                | 1.2              |
| American Airlines              | 1.2              |
| Aer Lingus                     | 1.2              |
| Turkish Airlines               | 1.3              |
| Lufthansa                      | 1.3              |
| British Airways                | 1.4              |
| Air France                     | 1.4              |
| KLM Royal Dutch Airlines       | 1.4              |
| TAP Air Portugal               | 1.4              |
| Iberia                         | 1.4              |
| Virgin Atlantic                | 1.4              |
| Air Canada                     | 1.4              |
| Emirates                       | 1.5              |
| Qatar Airways                  | 1.5              |
| Thai Airways                   | 1.5              |
| Vietnam Airlines               | 1.5              |
| Malaysia Airlines              | 1.5              |
| Japan Airlines (JAL)           | 1.6              |
| ANA (All Nippon Airways)       | 1.6              |
| Qantas                         | 1.7              |
| Cathay Pacific                 | 1.7              |
| Finnair                        | 1.7              |
| Swiss International Air Lines  | 1.8              |
| Austrian Airlines              | 1.8              |
| Etihad Airways                 | 1.8              |
| Korean Air                     | 1.9              |
| EVA Air                        | 1.9              |
| SAS (Scandinavian Airlines)    | 1.9              |
| Singapore Airlines             | 2.0              |
| Hawaiian Airlines              | 2.0              |
| China Airlines                 | 2.0              |
| LATAM Airlines                 | 2.0              |
| Avianca                        | 2.0              |
| South African Airways          | 2.0              |
| Philippine Airlines            | 2.0              |
| Asiana Airlines                | 2.1              |
| Garuda Indonesia               | 2.1              |
| Air New Zealand                | 2.1              |
| Oman Air                       | 2.1              |
| Royal Air Maroc                | 2.2              |
| Saudi Arabian Airlines (Saudia)| 2.2              |
| SriLankan Airlines             | 2.2              |
| Air India                      | 2.3              |
| Hainan Airlines                | 2.3              |
| China Southern Airlines        | 2.3              |
| China Eastern Airlines         | 2.3              |
| Gulf Air                       | 2.3              |
| Azul Brazilian Airlines        | 2.5              |
| Singapore Airlines Suites      | 3.0              |

---

## 2. Date (Season/Holiday) Modifiers

| Date       | Price Multiplier | Reason                             | Countries Affected     |
|------------|------------------|------------------------------------|------------------------|
| Jan 1      | 2.0              | New Year's Day surge pricing       | Global                 |
| Jan 2-5    | 1.6              | Post-New Year return travel        | Global                 |
| Jan 6      | 1.4              | Epiphany Holiday                   | ES, IT, DE, AT         |
| Jan 7-15   | 1.1              | Post-holiday lull                  | Global                 |
| Jan 16-31  | 1.0              | Off-peak winter pricing            | Global                 |
| Feb 1-10   | 1.0              | Low season, cheap fares            | Global                 |
| Feb 11-14  | 1.3              | Valentine’s Day trips              | Global                 |
| Feb 15-25  | 1.5              | School winter break                | DE, FR, UK, CH, NL     |
| Feb 26-28  | 1.0              | Low season resumes                 | Global                 |
| Mar 1-15   | 1.0              | Late winter shoulder season        | Global                 |
| Mar 16-19  | 1.2              | Early spring travel picks up       | Global                 |
| Mar 17     | 1.5              | St. Patrick’s Day                  | IE, US                 |
| Mar 20-24  | 1.3              | Spring travel picks up             | Global                 |
| Mar 25-31  | 1.5              | Easter/Spring Break starts         | Global                 |
| Apr 1-10   | 1.8              | Peak Easter holiday travel         | Global                 |
| Apr 11-20  | 1.3              | Easter return traffic              | Global                 |
| Apr 21-23  | 1.5              | Eid al-Fitr                        | SA, AE, IN, ID, MY, PK   |
| Apr 24-30  | 1.0              | Spring shoulder season             | Global                 |
| May 1      | 1.4              | May Day (Labor Day)                | EU, CN, RU, BR         |
| May 2-5    | 1.4              | Early summer trips                 | EU, CN, RU             |
| May 6-20   | 1.0              | Pre-summer lower demand            | Global                 |
| May 21-31  | 1.3              | Memorial Day (USA), early travel   | US                     |
| Jun 1-10   | 1.4              | Start of summer travel             | Global                 |
| Jun 11-20  | 1.5              | Peak pre-holiday travel            | Global                 |
| Jun 21-30  | 1.7              | Schools close, summer peak         | EU, US, CA, UK         |
| Jul 1-10   | 2.0              | Peak summer vacation season        | Global                 |
| Jul 4      | 1.7              | Independence Day                   | US                     |
| Jul 11-20  | 2.0              | High summer pricing continues      | Global                 |
| Jul 14     | 1.5              | Bastille Day                       | FR                     |
| Jul 21-31  | 1.9              | Mid-summer, still expensive        | Global                 |
| Aug 1      | 1.5              | Swiss National Day                 | CH                     |
| Aug 1-10   | 1.8              | Late summer vacations              | Global                 |
| Aug 15     | 1.6              | Assumption Day                     | IT, FR, ES, DE         |
| Aug 11-20  | 1.5              | Summer winding down                | Global                 |
| Aug 21-31  | 1.2              | Back-to-school, demand drops       | EU, US, UK             |
| Sep 1-10   | 1.0              | Shoulder season, cheaper fares     | Global                 |
| Sep 11-30  | 0.9              | Low demand, cheap flights          | Global                 |
| Oct 1-10   | 0.9              | Off-season continues               | Global                 |
| Oct 3      | 1.4              | German Unity Day                   | DE                     |
| Oct 11-20  | 1.0              | Fall travel picks up               | Global                 |
| Oct 31     | 1.3              | Halloween                          | US, UK, CA             |
| Nov 1-10   | 1.1              | Pre-holiday travel starts          | Global                 |
| Nov 11     | 1.2              | Veterans Day / Armistice Day       | US, FR, DE             |
| Nov 20-22  | 1.4              | Thanksgiving travel begins         | US                     |
| Nov 23-26  | 1.8              | Thanksgiving peak travel           | US                     |
| Nov 27-30  | 1.5              | Black Friday, Cyber Monday         | US, CA, UK             |
| Dec 1-10   | 1.3              | Christmas travel begins            | Global                 |
| Dec 11-20  | 1.6              | Pre-Christmas peak travel          | Global                 |
| Dec 21-24  | 2.2              | Christmas holiday peak             | Global                 |
| Dec 25     | 1.3              | Cheaper day to fly (low demand)      | Global                 |
| Dec 26-30  | 1.8              | Post-Christmas return travel       | Global                 |
| Dec 31     | 2.0              | New Year's Eve surge pricing       | Global                 |

## 3. Population Modifiers
| Population Range   | Example Cities (Approx.)         | Pop. Multiplier |
|--------------------|----------------------------------|-----------------|
| <10k               | Remote villages/islands          | 2.6             |
| 10k–49k            | Many small towns/cities          | 2.4             |
| 50k–99k            | St. Gallen (~75k), Truro (~60k)   | 2.2             |
| 100k–499k          | Aberdeen (~200k), Brest (~140k)  | 2.0             |
| 500k–999k          | Glasgow (~600k), Leeds (~790k)   | 1.8             |
| 1.0M–1.999M        | Vienna (~1.9M), Budapest (~1.8M) | 1.6             |
| 2.0M–2.999M        | Warsaw (~1.8M), Bucharest (~2.0M) | 1.5             |
| 3.0M–4.999M        | Berlin (~3.7M), Prague (~1.3M)   | 1.4             |
| 5.0M–6.999M        | Barcelona (~5.5M metro), Dallas (~6.5M metro) | 1.3  |
| 7.0M–9.999M        | Hong Kong (~7.4M), London (~9M)* | 1.2             |
| ≥10M               | Tokyo (~14M), Shanghai (~25M)     | 1.0             |


## 4. Flight-Frequency Modifier (FFM)
| Direct Flights per 5 Days | Price Multiplier | Notes                                          |
|---------------------------|------------------|------------------------------------------------|
| 100+                      | 0.8              | Ultra-high frequency routes                    |
| 50 – 99                   | 0.9              | Very frequent routes                           |
| 20 – 49                   | 1.1              | Frequent but not oversaturated                 |
| 10 – 19                   | 1.3              | Limited direct flight options                  |
| 5 – 9                     | 1.5              | Few flights, higher price due to scarcity      |
| 2 – 4                     | 1.8              | Very rare direct flights                       |
| 1                         | 2.0              | Only one flight available over 5 days          |
| 0                         | 2.5+             | No direct flights (layovers required)          |

## 5. Short-Notice Modifier (SNM)
| Days Before Departure   | Price Modifier | Explanation                                     |
|-------------------------|----------------|-------------------------------------------------|
| 90+ days                | 0.9            | Early bird deals; low demand                    |
| 60–89 days              | 1.0            | Typical booking window                          |
| 30–59 days              | 1.1            | Moderate demand                                 |
| 14–29 days              | 1.2            | Increased demand                                |
| 7–13 days               | 1.4            | Higher demand as departure nears                |
| 3–6 days                | 1.6            | Last-minute bookings, higher prices             |
| 1–2 days                | 1.8            | Very last minute; premium fares                 |
| Same day departure      | 2.0+           | Urgent travel; highest premium                  |


## 6. Aircraft Capacity Multiplier (ACM)

| Seat Capacity | Example Aircraft                              | ACM  | Notes                                               |
|---------------|-----------------------------------------------|------|-----------------------------------------------------|
| < 50          | Small turboprops (e.g., DHC-6 Twin Otter)      | 1.3  | High per-seat cost due to low capacity             |
| 50–100        | Regional jets (e.g., CRJ-900, Embraer E175)      | 1.1  | Moderate per-seat cost                              |
| 100–200       | Narrow-body jets (e.g., A319, B737)              | 1.0  | Standard fare                                       |
| 200–300       | Larger narrow-bodies (e.g., B737-800/Max, A321)  | 0.9  | Economies of scale reduce per-seat cost             |
| 300+          | Wide-bodies (e.g., B777, A350)                   | 0.8  | Lowest per-seat cost due to high capacity           |


## 7. Route Classification Multiplier (RCM)
| Route Classification      | Example Routes                      | Description                                       | RCM   |
|---------------------------|-------------------------------------|---------------------------------------------------|-------|
| Pure Business             | LHR ↔ FRA, JFK ↔ ORD                | High corporate travel; strong demand              | 1.3–1.5 |
| Mixed Business/Leisure    | LHR ↔ JFK, LAX ↔ SFO                | Both business and leisure travelers               | 1.1–1.2 |
| Pure Leisure              | LGW ↔ PMI (Palma), OSL ↔ AGP         | Vacation destinations; seasonal demand            | 1.0–1.1 |
| Essential/Remote          | EDI ↔ SYY, YYT ↔ YDF                 | Lifeline routes to remote areas                   | 1.2–1.4 |
| Low-Cost Tourist          | STN ↔ RYG                           | Ultra-budget routes with lower base fares         | 0.8–1.0 |
| Hub-to-Hub                | LHR ↔ FRA, CDG ↔ AMS                 | Major airline hubs with frequent connections      | 1.1–1.3 |
| Seasonal Charter          | European beach routes in summer, ski destinations in winter | Strong seasonal demand, charter pricing  | 1.0–1.2 |

