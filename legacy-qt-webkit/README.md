# Legacy Qt

This image is supposed to ease the pain of building legacy Qt applications.

To use it, volume mount the legacy project directories and run the relevant build commands from inside the container.

E.g.

```
docker pull hutchisont/legacy-qt

// from a dir containing legacy projects
docker run -ti --rm -v $(pwd):/usr/src hutchisont/legacy-qt bash

// now inside container
root@blah: cd <some_legacy_project> && qtchooser -qt=qt5 -run-tool=qmake . && make clean && make -j8 && make install
```
