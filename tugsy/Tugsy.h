#ifndef TUGSY_H
#define TUGSY_H

#include <thread>
#include <SDL2/SDL.h>

#include <sdl.h>

class Tugsy {
public:
    Tugsy(SdlState &state);
    int OnExecute();
    bool OnInit();
    void OnEvent(SDL_Event* Event);
    void OnLoop();
    void OnRender();
    void OnCleanup();

private:
    SdlState sdlState;
};

#endif
