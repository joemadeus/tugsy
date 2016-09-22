# tugsy
A surprise project for my fiancé, who likes to watch the tugboats in Providence Harbor.

This is the app and GIS data for a single-use computer that gathers vessel NEMA data from a radio in our house and from online sources, then displays that data on a map. (It's also my first foray into a full C++ app, as well as my first app with any sort of extensive graphics code.) SDL2 is used to render to the screen and to handle touch events. It uses existing code that tunes/controls an SDR and transforms the data into position info.

The hardware for the computer in question is this: 
https://www.olimex.com/Products/OLinuXino/A20/A20-OLinuXIno-LIME2-4GB/open-source-hardware

...and its screen/touch interface:
https://www.olimex.com/Products/OLinuXino/LCD/LCD-OLinuXino-7TS/open-source-hardware
