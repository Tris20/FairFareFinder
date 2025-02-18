Variables
BA = Base Price (in EUR/minute)
BA = 0.15 EUR/min

AM = Airline Multiplier
- Reflects typical airline cost level (e.g., 0.8 for easyJet, 1.6 for Loganair)

FD = Flight Duration (in minutes)
- e.g., 130 minutes for a 2 hr 10 min flight

PM = Population Modifier
- Based on origin & destination population brackets

DM = Date (Season/Holiday) Modifier
- E.g., 1.5 during winter or summer breaks

FFM = Flight-Frequency Modifier
- Based on how many direct flights exist in a 5-day span

SNM = Short-Notice Modifier
- Reflects how many days before departure the booking is made



PRICE=((AM)×(BA)×(FD))×(PM)×(DM)×(FFM)×(SNM)



# NOTES
- dont use day of week modifier: our result pice should reflect the 5 day window
- dont use time of day modifier: our result price shuold reflect the 5 day window
