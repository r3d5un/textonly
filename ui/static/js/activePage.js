document.querySelectorAll('.nav-link').forEach(link => {
    console.log(link.href);
    if (link.href === window.location.href) {
        link.classList.add('active');
    }
})
