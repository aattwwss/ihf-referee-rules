window.addEventListener('scroll', () => {
    const scrollToTopButton = document.getElementById('scroll-button');
    if (window.scrollY > 200) {
        // scrollToTopButton.classList.add('show');
        scrollToTopButton.style.display = 'flex';
        const opacity = Math.min(0.5, window.scrollY/2000);
        scrollToTopButton.style.opacity = `${opacity}`;
    } else {
        // scrollToTopButton.classList.remove('show');
        scrollToTopButton.style.display = 'none';
    }
});