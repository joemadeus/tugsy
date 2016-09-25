RESOURCES=res

EXECUTABLE=tugsy
CXX=g++
CXXFLAGS=-std=c++11 -Wall $(SDL_FLAGS)
INCLUDES=-I. -I./$(RESOURCES)
LIBS=-lSDL2_image
SDL_FLAGS=`sdl2-config --libs --cflags`
SOURCES=Tugsy.cpp sdl.cpp View.cpp PositionData.cpp

all: $(SOURCES) $(EXECUTABLE)

# g++ -o tug main.cpp sdl.cpp -std=c++11 -I. `sdl2-config --cflags --libs`
$(EXECUTABLE):
	$(CXX) -o $(EXECUTABLE) $(SOURCES) $(CXXFLAGS) $(INCLUDES) $(LIBS)

clean:
	rm -f *.o $(EXECUTABLE)
