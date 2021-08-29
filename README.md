# CS:GO WeakSpots
An EZ to use website that Generates Heatmaps through the CS:GO Demo files to Analyze your WeakSpots.


# Output Example
<img src="https://i.ibb.co/YPYX6tK/electronic-ancient.jpg">

* **Player Name:** &nbsp;&nbsp;electronic
* **Download:** &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Demo File(309MB)](https://drive.google.com/file/d/19DgALLPVG3eyENWfeFa6X-DNpp1J_i9l/view?usp=sharing)
* **Scoreboard:** &nbsp;&nbsp;&nbsp;&nbsp;[HLTV](https://www.hltv.org/stats/matches/mapstatsid/122163/natus-vincere-vs-faze)

# Hosted At
### [WeakSpots](https://weakspots.herokuapp.com/) ###
Try it out using:
* **Player Name:** &nbsp;&nbsp;tuxa20
* **Download:** &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;[Demo File(187MB)](https://drive.google.com/file/d/1DbWxnm-Hy8g39Upqbq2vN_4R8G6N32jc/view?usp=sharing)
* **Scoreboard:** &nbsp;&nbsp;&nbsp;&nbsp;[FACEIT](https://www.faceit.com/en/csgo/room/1-db6ad115-bce6-4b9f-973d-899bd0709b02/scoreboard)

NOTE: Heroku has a request timeout of 30sec, so bigger demo files will not get through :(

# About the Project
This project was made with Go(Golang) and Bootstrap.
It mainly uses the following packages and resources:  
* [demoinfocs-golang](https://github.com/markus-wa/demoinfocs-golang)  
* [go-heatmap](https://github.com/dustin/go-heatmap)
* [csgo-overviews](https://github.com/zoidbergwill/csgo-overviews)

# Limitations
It is not optimized to provide multi-level heatmaps for maps like Nuke, Vertigo, etc.