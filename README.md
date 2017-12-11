# tugsy
A surprise project for my ~fiancé~ wife, who likes to watch the tugboats in Providence Harbor.

This is the app and GIS data for a single-use computer that gathers vessel NEMA data from a radio in our house and from online sources, then displays that data on a map. SDL2 (see https://github.com/veandco/go-sdl2/) is used to render to the screen and to handle touch events, and `aislib` (https://github.com/andmarios/aislib) is used to decode NEMA messages.

There are two ways to interact with the app: by touching the screen, which displays information for the vessel under the touch; or by pressing a hardware button on the front of the machine's frame, which cycles through three pre-set views of PVD's inner harbor, from PVD to Bristol, and from PVD to the shipping lanes in Long Island Sound off Point Judith.

All of the data for land-based features come from OpenStreetMap data exports (see https://www.openstreetmap.org/#map=10/41.6493/-71.5876); all of the data for maritime features comes from *our tax dollars at work*, a.k.a., NOAA Electronic Navigational Charts (see http://www.charts.noaa.gov/InteractiveCatalog/nrnc.shtml).

The mainboard is this: https://www.olimex.com/Products/OLinuXino/A20/A20-OLinuXIno-LIME2-4GB/open-source-hardware

...and its screen/touch interface: https://www.olimex.com/Products/OLinuXino/LCD/LCD-OLinuXino-7TS/open-source-hardware

...and the SDR: https://www.nooelec.com/store/sdr/sdr-receivers/nesdr-smart-xtr-sdr.html

The enclosure for this is partly ready-made frame (see https://www.olimex.com/Products/OLinuXino/LCD/LCD7-METAL-FRAME/) and partly laser-cut pieces of my own design -- that frame is useful but ugly.

I'm building a quarter-wave groundplane antenna, as well, with the generous help of @farzadb82 -- the model for it and various other bits can be found in `misc`.

<a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/"><img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-nc-sa/4.0/88x31.png" /></a><br />This work is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by-nc-sa/4.0/">Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License</a>.
