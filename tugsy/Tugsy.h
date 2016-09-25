#ifndef TUGSY_H
#define TUGSY_H

#include <vector>
#include <stdio.h>
#include <thread>
#include <signal.h>
#include <SDL2/SDL_events.h>

#include "sdl.h"
#include "PositionData.h"
#include "View.h"

using namespace std;

class Tugsy {
public:
    Tugsy(SdlState &state, PositionData &positions);
    int onExecute();

private:
    void onEvent(SDL_Event* Event);
    bool onInit();
    void onLoop();
    void onRender();
    void onCleanup();

    View& currentView();

    SdlState sdlState;
    PositionData positionData;
    std::vector<View> views;
    unsigned int currentViewIndex = 0;
    long lastUpdateMillis = 0;
};

bool running = true;

void sigHandler(int param);

#endif
