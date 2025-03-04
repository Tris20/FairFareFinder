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

LFM = Load Factor Modifier 
- a factor which approximates how busy the plan will be by combining snm, ffm, pm, and dm 

FINAL PRICE=((AM)×(BA)×(FD))×(PM)×(DM)×(FFM)×(SNM)x(LFM)


LFM
LFM=1+(LF−0.5)×γ

LF 
LF=0.5+α1(SNM−1)−α2(FFM−1)+α3(PM−1)+α4(DM−1)
where 
 - 0.5: Baseline load factor (i.e., 50% full if everything was “normal”).
 - α: Weights that tell us how much each variable pushes/pulls LF from that 0.5 baseline. 
α1 =0.15 (short notice is a strong indicator),
𝛼2=0.10 (flight frequency has a moderate impact),
𝛼3=0.10 (population size also moderate),
𝛼4=0.10 (date factor moderate).

eg 
LF=0.5+0.15×(SNM−1)−0.10×(FFM−1)+0.10×(PM−1)+0.10×(DM−1).
This will typically produce an LF between ~0.4 and ~1.5 for most real scenarios.

# NOTES
- dont use day of week modifier: our result pice should reflect the 5 day window
- dont use time of day modifier: our result price shuold reflect the 5 day window
