# Test group: CHE

This test group defines sets of GeoZones to test the compliance of a Geo-Awareness service as defined in Article 9 of
[the U-space regulation](https://eur-lex.europa.eu/legal-content/EN/TXT/HTML/?uri=CELEX:32021R0664&from=EN#d1e1006-161-1).

## CHE-1 Test Scenario

[CHE-1 Test Scenario](./geo-awareness-che-1.json) contains 4 types of geometry types:

- Circle (single per feature)
- Polygon (single per feature)
- Circles (multiple circles in one feature)
- Polygons (multiple polygons in one feature)

It provides altitude references in both AGL and AMSL in FT and M per ED-269 Spec.
It has schedules associated with features that govern when applicable.
Specific features applicable to certain U-space classes only. 
It uses mock Authority information for USSPs wishing to use data for their users.

| Test Case Identifier       | Brief Description                                                                                                                                                        | 
|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Montreux Concert Area      | Sample circular restriction that only applies on Saturday and Sundays with mixed references for lower and upper limits (AGL and AMSL in M)                               | 
| Flugplatz Reichenbach      | Sample concentric circular restriction with a validity period all in AGL (M)                                                                                             |
| Lausanne Airport           | Sample multiple polygon restrictions active all the time in AGL (M)                                                                                                      |
| Montreux Wildlife Preserve | Sample single polygon alerting to need for authorisation in AGL(M)                                                                                                       |
| Gantrisch Nature Park      | Sample polygon restrictions that:<br>- Provides an informational message for Open Category and<br>- Alerts to need for authorisation for Specific and Certified Category |

All lower and upper limits in AGL (FT)

The requirements coverage of this scenario is analyzed in this [document](https://docs.google.com/spreadsheets/d/1NIlRHtWzBXOyJ58pYimhDQDqsEyToTQRu2ma3AYXWEU/edit?usp=sharing). 
