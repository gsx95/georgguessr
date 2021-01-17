import json

capitals_file = 'capitals.json'
cities_file = 'cities_pop.json'
country_codes_file = 'country-codes.json'

name_to_code = {}
countries = {}

with open(country_codes_file) as f:
    data = json.load(f)
    for country in data:
        name_to_code[country["name"]] = country["alpha-2"]
        countries[country["alpha-2"]] = {
            "country": country["name"],
            "cities": [],
            "capital": ""
        }

with open(capitals_file) as f:
    data = json.load(f)
    capitals = data["features"]
    for capital in capitals:
        country_code = capital["properties"]["iso2"]
        if country_code not in countries:
            continue
        if "city" in capital["properties"]:
            capital_name = capital["properties"]["city"]
            countries[country_code]["capital"] = capital_name

with open(cities_file) as f:
    data = json.load(f)
    for city in data:
        name = city["fields"]["city"]
        country_code = city["fields"]["country"].upper()
        pop = city["fields"]["population"]
        if country_code in countries:
            city = {
                "name": name,
                "population": pop,
            }
            countries[country_code]["cities"].append(city)

countries_json = json.dumps(countries, indent = 4)
f = open("countries.json", "a")
f.write(countries_json)
f.close()

codes_json = json.dumps(name_to_code, indent = 4)
x = open("codes.json", "a")
x.write(codes_json)
x.close()