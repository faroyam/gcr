const chat = document.querySelector(".chatfield");
const msg = document.querySelector(".msg");
const counter = document.querySelector(".usercounter");

const sendbtn = document.querySelector(".sendbtn");

sendbtn.addEventListener("click", async () => {
    if (msg.value != "") await send(msg.value);
});

document.addEventListener("keydown", event => {
    if (event.key !== "Enter") return;
    sendbtn.click();
});
