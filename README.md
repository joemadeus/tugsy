# tugsy
A surprise project for my fiancé, who likes to watch the tugboats in Providence Harbor.

This is the app and GIS data for a single-use computer that gathers vessel NEMA data from a radio in our house and from online sources, then displays that data on a map. (It's also my first foray into a full C++ app, as well as my first app with any sort of extensive graphics code.) SDL2 is used to render to the screen and to handle touch events. It uses existing code that tunes/controls an SDR and transforms the data into position info.

There are two ways to interact with the app: the first, by touching the screen, will (eventually, at least) display information for the vessel under the touch; the second, by pressing a hardware button on the machine's frame, cycles through three pre-set views, one for PVD's inner harbor; from PVD to Bristol; and from PVD to the shipping lanes in Long Island Sound off Point Judith.

All of the data for land-based features come from OpenStreetMap data exports (see https://www.openstreetmap.org/#map=10/41.6493/-71.5876); all of the data for maritime features comes from *our tax dollars at work*, a.k.a., NOAA Electronic Navigational Charts, or ENCs (see http://www.charts.noaa.gov/InteractiveCatalog/nrnc.shtml).

The mainboard for the computer in question is this: https://www.olimex.com/Products/OLinuXino/A20/A20-OLinuXIno-LIME2-4GB/open-source-hardware

...and its screen/touch interface: https://www.olimex.com/Products/OLinuXino/LCD/LCD-OLinuXino-7TS/open-source-hardware

...and the SDR: http://www.nooelec.com/store/sdr/nesdr-mini-rtl2832-r820t.html

The enclosure for this is partly ready-made frame (see https://www.olimex.com/Products/OLinuXino/LCD/LCD7-METAL-FRAME/) and partly laser-cut pieces of my own design. That frame is useful but ugly.

### Code organization
'Tugsy' is the main class and contains the main method. Its event loop prompts the radio and other sources for data, feeding them into a class that organizes and selects position info for display. The control loop then pulls the coallated data and feeds it to a view class, which figures out which bits of screen need to be updated, then does so by writing to a class that manages the SDL context. The display is not updated very often, usually five frames per second or whenever one of the events described earlier has to be handled (which is also handled by Tugsy's main event loop.)
