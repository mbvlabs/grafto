const navToggle = document.querySelector('.nav-toggle');
const navMenu = document.querySelector('.nav-menu');

navToggle.addEventListener('click', () => {
	navMenu.classList.toggle('nav-menu-active');
	navToggle.classList.toggle('nav-toggle-active');
});
