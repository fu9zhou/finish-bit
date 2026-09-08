# Fixed lunar calendar data

FinishBit only exposes Gregorian/lunar date conversion for input years
1900–2100. It does not need the astronomy, solar terms, festivals or other
calendar features of the former `github.com/6tail/lunar-go` dependency.

`../../lunar_table.go` contains 202 year records, 808 bytes as `uint32` values.
The preceding lunar year 1899 is required because 1900-01-01 maps to lunar
1899-12-01. The complete lunar year 2100 ends on Gregorian 2101-01-28;
`calendar.solar` continues to allow that result. Input-year restrictions are
unchanged. These are civil dates, not calculations based on the host timezone.

## Provenance and representation

The table was evaluated on 2026-09-08 using the previously pinned
[lunar-go v1.4.6](https://github.com/6tail/lunar-go/tree/v1.4.6). It records
calendar results, not a copied astronomical implementation. Attribution and
the existing MIT license copy remain in `THIRD_PARTY_NOTICES.md` and
`third_party/licenses/`; the Go module is no longer needed to build or test.

For each year from 1899 through 2100, `NewLunarYear(year).GetMonths()` was
filtered to months whose `GetYear()` equals that year. The low four bits are
the absolute value of the negative month number, or zero when no leap month
exists. Starting at bit four, each chronological month contributes a bit set
for 30 days, clear for 29 days. The leap month appears immediately after the
regular month with the same number. All month starts were checked for
continuity using `GetFirstJulianDay()` converted to a Gregorian date.
The epoch is Gregorian 1899-02-10, lunar 1899-01-01.

FinishBit's own conversion implementation adds or subtracts these month/year
lengths and validates leap months and day bounds explicitly. It does not copy
the upstream date conversion functions or use panic recovery for validation.

## Compatibility and independent checks

Before removing the dependency, both complete supported date domains were
evaluated with v1.4.6 to produce two SHA-256 reference digests:

- Gregorian 1900-01-01 through 2100-12-31: 73,414 dates; each line is
  `YYYY-MM-DD lunarYear signedLunarMonth lunarDay ChineseText\n`.
  Digest: `d7065ef43ad2765a1a95f4c154c46cfbd79cd478c43a783571a55405a2b747bc`.
- All valid lunar dates in years 1900–2100: 73,412 dates, in chronological
  month order; each line is `lunarYear signedLunarMonth lunarDay YYYY-MM-DD\n`.
  Digest: `0b39f730429d71be47e4b818838b71b30853c0f7674bc5a7ee551298679c4af0`.

The signed month is negative for leap months. Text uses UTF-8, fields are
separated by single ASCII spaces, integers have no padding, and lines end in
LF. The permanent tests calculate the same digests from our implementation;
they do not derive expected results from the table and require no network or
external module. Invalid leap months, 30th days of short months and endpoint
years have additional explicit tests.

Independent fixtures check Lunar New Year, a Gregorian leap day, and the
2033 leap eleventh month against the Hong Kong Observatory's
[2024](https://www.hko.gov.hk/en/gts/time/calendar/text/files/T2024e.txt),
[2033](https://www.hko.gov.hk/en/gts/time/calendar/text/files/T2033e.txt) and
[2034](https://www.hko.gov.hk/en/gts/time/calendar/text/files/T2034e.txt)
tables. This is a spot-check, not a claim of full agreement with every HKO
date. The [HKO explanation](https://www.hko.gov.hk/en/gts/time/conversion.htm)
notes uncertainty for some distant future new moons near midnight. The
fixed table deliberately preserves the former library's outputs; extending
the supported interval or revising calendar conventions requires a new data
review, independent checks and an explicit compatibility decision.
