# Skookum's Facebot?

The beginnings of a web app that can recognize Skookumites. Currently, there is a lot
of manual setup.

## Running

You must have the Facebox docker image running locally. TO do that, you'll need to
get a key from [machinebox.io](http://machinebox.io) which is FREE:

```
MB_KEY="YOUR MB key" docker run -p 8080:8080 -e "MB_KEY=$MB_KEY" machinebox/facebox
```

Then you need to train the Facebox with pictures (I was using the Bamboo HR pics)
and assign it a Name. The Name is used to lookup the sound file. You can upload photos
to the facebox by going to `http://localhost:8080` and clicking on `Try it now` it the
"Teach faces" section. That will show a form that has a ID, Name, and URL field.
Put the URL to the image (you can paste in the URL to the Bamboo image) and a name and
click `Post`.

The `/sound` directory has the mp3s of the sounds. Each sound needs to be the
lowercase version of the name.mp3.  I used espeakbox to synthesize text-to-speech.
Espeakbox offers another docker image, so:

```
docker run --name espeakbox -d -p 8082:8080 parente/espeakbox

curl http://localhost:8082/speech\?text\=Hello%20Enrique > sound/enrique.mp3
```

Once that's done, you can run this app, which starts a web server:

```
> go run main.go (or build it and run it)
```

Then, goto `http://localhost:8081/static` and put your face in the video window. It
should say "Hello <your name>"
