// app.js will hook the websocket and update the feed in realtime
// Author: Dylan Evans|fin3ss3g0d
window.addEventListener('focus',()=>{
    clearNotifications();
})

let muted = true;
let capcount = 0;

$('#audio-control').click(function(){
    if( $(this).text() === 'Unmute' ) {
        $(this).text('Mute');
        muted = false;
    } else {
        $(this).text('Unmute');
        muted = true;
    }
});

function showTokens (count) {
    var tokens = document.getElementById(count);
    if (tokens.style.display === "inline-block") {
        tokens.style.display = "none";
    } else {
        tokens.style.display = "inline-block";
    }
}

ws = new WebSocket("ws://localhost:1337/ws");
ws.onopen = function(evt) {
    console.log("OPEN");
}
ws.onclose = function(evt) {
    console.log("CLOSE");
    ws = null;
}
ws.onerror = function(evt) {
    console.error("ERROR: ", evt);
}

const feedDiv = document.getElementById('feed');

ws.onmessage = function(evt) {
    //console.log("RESPONSE: " + evt.data);
    let parsed = JSON.parse(evt.data)
    let templ = "";
    if (parsed.event == "Captured Session") {
        //console.log('Tokens: ' + parsed.tokens)
        templ = `
        <div class="box">
        <article class="media">
            <div class="media-content">
            <div class="content">
            <p>
                <strong>` + parsed.event + `</strong>
                <small>` + parsed.time + `</small> <br />
                ` + parsed.message + `
            </p>
            <button class="token-button" onclick="showTokens('token-content` + capcount.toString() + `')">View Tokens</button> 
            <div id="token-content` + capcount.toString() + `" class="token-space">` + `<br>` + parsed.tokens + `</div>
            </div>
            </div>
        </article>
        </div>
    `;
    capcount += 1;
    } else {
        templ = `
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
    }

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