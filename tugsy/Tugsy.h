#ifndef TUGSY_H
#define TUGSY_H

#include <stdio.h>
#include <thread>
#include <signal.h>
#include <SDL2/SDL.h>

#include <sdl.h>

using namespace std;

class Tugsy {
public:
    Tugsy(SdlState &state);
    int onExecute();
    bool onInit();
    void onEvent(SDL_Event* Event);
    void onLoop();
    void onRender();
    void onCleanup();

private:
    SdlState sdlState;
};

bool running = true;

void sigHandler(int param);

#endif
