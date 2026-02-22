// Animated tab title like oxal.org - Performance Optimized
function initAnimatedTitle() {
    const baseTitle = "uroboro";
    const fullText = "uroborouroborouroborouroborouroborouroborouroboro";
    let position = 0;
    let direction = 1;
    let titleIntervalId;

    function updateTitle() {
        // Only update if page is visible
        if (document.visibilityState !== 'visible') return;

        // Create a sliding window effect
        const windowSize = 12; // Number of characters to show
        let displayText;

        if (position + windowSize >= fullText.length) {
            // When we reach the end, show the base title
            displayText = baseTitle;
            position = 0;
        } else {
            // Show sliding window of the full text
            displayText = fullText.substring(position, position + windowSize);
            position += direction;
        }

        document.title = displayText;
    }

    // Start title animation with longer interval (300ms instead of 200ms)
    titleIntervalId = setInterval(updateTitle, 300);

    // Pause title animation when page is hidden
    document.addEventListener('visibilitychange', function() {
        if (document.visibilityState === 'hidden') {
            clearInterval(titleIntervalId);
            document.title = baseTitle; // Reset to base title when hidden
        } else {
            titleIntervalId = setInterval(updateTitle, 300);
        }
    });
}

// Spectacular uroboro Animation - Performance Optimized
function initUroboroAnimation() {
    const allLetters = document.querySelectorAll('.uroboro-letter');

    // Counterclockwise rotation animation for the entire circle - like a spinning record!
    anime({
        targets: '.uroboro-circle',
        rotate: '-360deg',
        duration: 12000, // Slightly faster for more dynamic feel
        easing: 'linear',
        loop: true
    });

    // Performance-optimized color inversion with page visibility and reduced frequency
    let inversionAnimationId;
    let lastUpdateTime = 0;
    const UPDATE_INTERVAL = 100; // Reduced from 50ms to 100ms (10fps instead of 20fps)

    function updateInversion(currentTime) {
        // Only update if page is visible and enough time has passed
        if (document.visibilityState === 'visible' && (currentTime - lastUpdateTime) >= UPDATE_INTERVAL) {
            const container = document.querySelector('.uroboro-container');
            if (!container) return;

            const containerRect = container.getBoundingClientRect();
            const centerX = containerRect.left + containerRect.width / 2;

            allLetters.forEach(letter => {
                const letterRect = letter.getBoundingClientRect();
                const letterCenterX = letterRect.left + letterRect.width / 2;

                // Invert color when letter is in the right half (under the mask)
                if (letterCenterX > centerX) {
                    letter.style.color = '#ffffff';
                    letter.style.textShadow = '0 0 12px rgba(255, 255, 255, 0.8), 0 0 6px rgba(255, 255, 255, 0.6)';
                } else {
                    letter.style.color = 'var(--primary)';
                    letter.style.textShadow = 'none';
                }
            });

            lastUpdateTime = currentTime;
        }

        inversionAnimationId = requestAnimationFrame(updateInversion);
    }

    // Start the animation loop
    inversionAnimationId = requestAnimationFrame(updateInversion);

    // Pause animation when page is hidden to save resources
    document.addEventListener('visibilitychange', function() {
        if (document.visibilityState === 'hidden') {
            cancelAnimationFrame(inversionAnimationId);
        } else {
            lastUpdateTime = 0; // Reset timer when page becomes visible again
            inversionAnimationId = requestAnimationFrame(updateInversion);
        }
    });

    // Add subtle pulsing to the mask
    anime({
        targets: '.inversion-mask',
        opacity: [0.8, 1, 0.8],
        duration: 3000,
        easing: 'easeInOutSine',
        loop: true
    });

    // Add floating animation to the whole container (more subtle)
    anime({
        targets: '.uroboro-container',
        translateY: [-3, 3, -3],
        duration: 5000,
        easing: 'easeInOutSine',
        loop: true
    });
}

// Enhanced reduced motion detection
function respectsReducedMotion() {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// Main initialization
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the animated title
    initAnimatedTitle();

    // Initialize the spectacular uroboro animation
    initUroboroAnimation();

    // Initialize dynamic tickertape
    initTickertape();

    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });

    // Add subtle animations on scroll
    const observerOptions = {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, observerOptions);

    // Observe sections for scroll animations
    document.querySelectorAll('section').forEach(section => {
        section.style.opacity = '0';
        section.style.transform = 'translateY(20px)';
        section.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(section);
    });

    // Initialize first section as visible
    const firstSection = document.querySelector('section');
    if (firstSection) {
        firstSection.style.opacity = '1';
        firstSection.style.transform = 'translateY(0)';
    }
});

// Dynamic Tickertape Generation - Simple Always-Visible Scroll
function initTickertape() {
    function generateRepeatingText(baseText, containerWidth, multiplier = 3) {
        // Calculate how many repetitions we need to fill viewport width
        // This ensures seamless scrolling with no gaps
        const charWidth = 8; // Approximate character width in pixels for the mono font
        const baseLength = baseText.length * charWidth;
        const totalNeeded = containerWidth * multiplier; // Default 3x viewport for seamless loop
        const repetitions = Math.ceil(totalNeeded / baseLength);

        return baseText.repeat(repetitions);
    }

    // Generate content for both tickertapes
    const viewportWidth = window.innerWidth;
    const topRepeatingText = generateRepeatingText('uroboro', viewportWidth, 3);
    const bottomRepeatingText = generateRepeatingText('uroboro', viewportWidth, 3); // Match top multiplier

    // Update both top and bottom tickertapes
    const topContent = document.querySelector('.tickertape-top .tickertape-content');
    const bottomContent = document.querySelector('.tickertape-bottom .tickertape-content');

    if (topContent) {
        topContent.textContent = topRepeatingText;
    }

    if (bottomContent) {
        bottomContent.textContent = bottomRepeatingText;
    }

    // Add ouroboros click behavior - bottom tickertape scrolls to top
    const bottomTickertape = document.querySelector('.tickertape-bottom');
    if (bottomTickertape) {
        bottomTickertape.addEventListener('click', function() {
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });
        bottomTickertape.style.cursor = 'pointer';
        bottomTickertape.title = 'Click to return to the beginning...';
        bottomTickertape.style.pointerEvents = 'auto';
    }

    // Top tickertape is not clickable - avoid annoying misclicks
    const topTickertape = document.querySelector('.tickertape-top');
    if (topTickertape) {
        topTickertape.style.cursor = 'default';
        topTickertape.style.pointerEvents = 'none';
    }
}

// Regenerate on window resize to maintain seamless scrolling
window.addEventListener('resize', initTickertape);