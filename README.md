# WelcomerImages
Image generation for Welcomer

# Example Payload

```
POST /images

{
    "filesize_limit": 8000000,
    "options": {
        "text": "Welcome ImRock\nyou are the 2424th member!",
        "guild_id": 341685098468343822,
        "user_id": 143090142360371200,
        "avatar": "a_70444022ea3e5d73dd00d59c5578b07e",
        "allow_gif": false,
        "layout": 2,
        "profile_alignment": 0,
        "background": "lodge",
        "font": "Inter-Bold",
        "border_colour": "471F2F",
        "border_width": 16,
        "text_alignment_x": 0,
        "text_alignment_y": 0,
        "profile_border_colour": "ffffff",
        "profile_border_width": 0,
        "profile_border_curve": 1,
        "text_stroke": 2,
        "text_stroke_colour": "471F2F",
        "text_colour": "471F2F"
    }
}
```

# Example response

`POST /images` will return the image directly if its size is less than
`filesize_limit` or `force_cache` has been set to true. If it is larger
than the limit or you have forced the cache the response will be:

```
{
    "success":true,
    "bookmarkable":"http://localhost:4200/images/q-EV6w9tAKeGC5CyHDYY-A.png",
    "image_data":{
        "id":"q-EV6w9tAKeGC5CyHDYY-A",
        "guild_id":341685098468343822,
        "size":356523,
        "path":"q-EV6w9tAKeGC5CyHDYY-A.png","expires_at":"2021-04-19T21:31:24.514714Z","created_at":"2021-04-12T21:31:24.514714Z"
    }
}
```