#ifndef SDL_H
#define SDL_H

#include <exception>
#include <iostream>
#include <utility>
#include <SDL2/SDL.h>
#include <SDL2/SDL_image.h>

using namespace std;

const int SCREEN_X = 0;
const int SCREEN_Y = 0;
const int SCREEN_WIDTH  = 480;
const int SCREEN_HEIGHT = 800;

const char PATH_SEP = '/';

class SdlState {
public:
	SdlState();
	~SdlState();

	bool addTex(SDL_Texture* texture);

	void paint(SDL_Texture* texture,
               const SDL_Rect* srcrect,
               const SDL_Rect* dstrect);
	/**
	 * Display the contents of the back buffer
	 */
	void flip();

	SDL_Texture* initTexture(const std::string &viewName, const std::string &resource);

private:
	void initContext();
	std::string getResource(const std::string &viewName, const std::string &resource);

	SDL_Window *window = NULL;
	SDL_Renderer *renderer = NULL;
	std::string baseRes;
};

void sdlError(const std::string &msg);

/*
 * Recurse through the list of arguments to clean up, cleaning up
 * the first one in the list each iteration.
 */
template<typename T, typename... Args>
void cleanup(T *t, Args&&... args){
	//Cleanup the first item in the list
	cleanup(t);
	//Recurse to clean up the remaining arguments
	cleanup(std::forward<Args>(args)...);
}

/*
 * These specializations serve to free the passed argument and also provide the
 * base cases for the recursive call above, eg. when args is only a single item
 * one of the specializations below will be called by
 * cleanup(std::forward<Args>(args)...), ending the recursion
 * We also make it safe to pass nullptrs to handle situations where we
 * don't want to bother finding out which values failed to load (and thus are null)
 * but rather just want to clean everything up and let cleanup sort it out
 */
template<>
inline void cleanup<SDL_Window>(SDL_Window *win){
	if (!win){
		return;
	}
	SDL_DestroyWindow(win);
}
template<>
inline void cleanup<SDL_Renderer>(SDL_Renderer *ren){
	if (!ren){
		return;
	}
	SDL_DestroyRenderer(ren);
}
template<>
inline void cleanup<SDL_Texture>(SDL_Texture *tex){
	if (!tex){
		return;
	}
	SDL_DestroyTexture(tex);
}
template<>
inline void cleanup<SDL_Surface>(SDL_Surface *surf){
	if (!surf){
		return;
	}
	SDL_FreeSurface(surf);
}

#endif
