

document.addEventListener("DOMContentLoaded", function(){
    document.querySelector("#shorten").addEventListener("click", shortened)
})


async function shortened(){
    url = document.querySelector("#urlInput").value
    response = await fetch("/shorten", {
        method: "POST",
        body: new URLSearchParams({ url }),
        headers: { "Content-Type": "application/x-www-form-urlencoded" }
    });

    const result = await response.json();

    if (result.short_url) {
        document.getElementById("shortenedLink").innerHTML = `Short URL: <a href="${result.short_url}" target="_blank">${result.short_url}</a>`;
    } else {
        document.getElementById("shortenedLink").innerText = `Error: ${result.error}`;
    }
}

