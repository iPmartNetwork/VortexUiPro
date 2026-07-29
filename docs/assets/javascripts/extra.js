// ═══════════════════════════════════════════════════════════════════
// VortexUiPro — Custom JavaScript for MkDocs Material
// ═══════════════════════════════════════════════════════════════════

// ─── Smooth scroll for anchor links ───────────────────────────────
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const href = this.getAttribute('href');
            if (href === '#') return;
            const target = document.querySelector(href);
            if (target) {
                e.preventDefault();
                target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }
        });
    });
});

// ─── Copy button enhancement ───────────────────────────────────────
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('.highlight button.md-clipboard').forEach(btn => {
        btn.addEventListener('click', function () {
            const original = this.querySelector('.tooltip');
            if (original) {
                original.textContent = 'Copied!';
                setTimeout(() => {
                    original.textContent = 'Copy to clipboard';
                }, 2000);
            }
        });
    });
});

// ─── Active nav highlight on scroll ───────────────────────────────
document.addEventListener('DOMContentLoaded', function () {
    const observers = [];
    document.querySelectorAll('.md-content h2, .md-content h3').forEach(heading => {
        const id = heading.getAttribute('id');
        if (!id) return;

        const link = document.querySelector(`.md-nav__link[href="#${id}"]`);
        if (!link) return;

        const observer = new IntersectionObserver(entries => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    document.querySelectorAll('.md-nav__link--active').forEach(el => {
                        el.classList.remove('md-nav__link--active');
                    });
                    link.classList.add('md-nav__link--active');
                }
            });
        }, { rootMargin: '-80px 0px -80% 0px' });

        observer.observe(heading);
        observers.push(observer);
    });
});
