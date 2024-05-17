<div align="center">
<img height="150px" src="https://raw.githubusercontent.com/wisdom-oss/brand/main/svg/standalone_color.svg">
<h1>NLWKN Water Level Crawler</h1>
<h3>crawler-nlwkn-water-levels</h3>
<p>📏 crawler for water levels reported by the nlwkn's masurement stations</p>
<img src="https://img.shields.io/github/go-mod/go-version/wisdom-oss/crawler-nlwkn-water-levels?style=for-the-badge" alt="Go Lang Version"/>
</div>

> [!IMPORTANT]
> The data collected by this application has been licensed under the 
> [Datenlizenz Deutschland — Namensnennung — Version 2.0].
> When using the data please respect the license.
> More information [here].
> 
> [Datenlizenz Deutschland — Namensnennung — Version 2.0]: https://www.govdata.de/dl-de/by-2-0
> [here]: https://www.nlwkn.niedersachsen.de/opendata/nlwkn-daten-wichtige-informationen-zu-geodaten-und-anderen-datenbestanden-aus-unseren-aufgabenbereichen-196027.html

This crawler regularly accesses the [overview page] for the groundwater levels
reported by the [NLWKN].
It collects the information about the available stations and their reported
water levels and stores them in different tables (one for the stations, one for
the reported levels).

[overview page]: https://www.grundwasserstandonline.nlwkn.niedersachsen.de/Messwerte
[NLWKN]:  https://www.nlwkn.niedersachsen.de
