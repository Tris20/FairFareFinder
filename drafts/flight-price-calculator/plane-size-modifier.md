
| Seat Capacity | Example Aircraft             | ACM  | Notes                                      |
|---------------|------------------------------|------|--------------------------------------------|
| < 20          | Very small turboprops       | 1.5  | Tiny planes → very high cost per seat      |
| 20–50         | ATR 42, DHC Twin Otter      | 1.3  | Regional props, still expensive per seat   |
| 50–100        | CRJ-900, Embraer E-175      | 1.1  | Regional jets, moderate cost               |
| 100–200       | A319, B737-700, E-195       | 1.0  | Standard narrow-body short/medium-haul     |
| 200–300       | B737-800, A321, B757        | 0.9  | Larger narrow-bodies → slight savings      |
| 300+          | B777, A350, etc.            | 0.8  | Wide-bodies, more efficient per seat       |



Method B: Formula Based on Seats
If you don’t want brackets, you can use a simple function:

ACM=𝑘/sqrt(Seats)

Where 
𝑘 is chosen so that:

ACM ~ 1.0 for a mid-range capacity (say, 150 seats).
ACM > 1.0 for small planes.
ACM < 1.0 for large aircraft.
Example: 
If 𝑘=12 then for 150 seats:

ACM150
=12/sqrt(150)
≈12/12.25
≈0.98
≈1.0

Example2:
For a 50-seat turboprop:

ACM50
≈12/7.07
≈1.70

For a 300-seat wide-body:
ACM300
≈12/17.32
≈0.69

You can tune 
𝑘
k to get the multipliers you want.
