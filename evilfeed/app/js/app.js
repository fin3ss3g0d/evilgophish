// app.js will hook the websocket and update the feed in realtime
// Author: Dylan Evans|fin3ss3g0d
(function() {
    window.addEventListener('focus',()=>{
        clearNotifications();
    })

    let muted = true;

    $('#audio-control').click(function(){
        if( $(this).text() === 'Unmute' ) {
            $(this).text('Mute');
            muted = false;
        } else {
            $(this).text('Unmute');
            muted = true;
        }
    });

    ws = new WebSocket("ws://localhost:1337/ws");
    ws.onopen = function(evt) {
        console.log("OPEN");
    }
    ws.onclose = function(evt) {
        console.log("CLOSE");
        ws = null;
    }

    const feedDiv = document.getElementById('feed');

    ws.onmessage = function(evt) {
        //console.log("RESPONSE: " + evt.data);
        let parsed = JSON.parse(evt.data)
        const templ = `
            <div class="box">
            <article class="media">
                <div class="media-content">
                <div class="content">
                <p>
                    <strong>` + parsed.event + `</strong>
                    <small>` + parsed.time + `</small> <br />
                    ` + parsed.message + `
                </p>
                </div>
                </div>
            </article>
            </div>
        `;
        const template = Handlebars.compile(templ);
        const html = template(evt);
        const divElement = document.createElement('div');
        divElement.innerHTML = html;
        feedDiv.appendChild(divElement);
    
        addNotification("event");
        if (muted === false) {
            var audio = new Audio('notify.mp3');
            audio.play();
        }
    }
    ws.onerror = function(evt) {
        console.log("ERROR: " + evt.data);
    }

    let notifications = [];

    function addNotification (notification) {
        notifications.push(notification);
        showNotificationCount(notifications.length);
    }

    function clearNotifications () {
        notifications = [];
        showNotificationCount(notifications.length);
    }

    function showNotificationCount (count) {
        const pattern = /^\(\d+\)/;

        if (count === 0 || pattern.test(document.title)) {
            document.title = document.title.replace(pattern, count === 0 ? "" : "(" + count + ")");
        } else {
            document.title = "(" + count + ") " + document.title;
        }
    }
})();    