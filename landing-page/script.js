// Voice examples for each style
const voiceExamples = {
    professional: {
        title: "Professional Conversational",
        content: `"Today I discovered something interesting about WebSocket lifecycle management. While debugging what seemed like a simple memory issue, I uncovered a subtle but important pattern about event listener cleanup that every real-time application developer should know. 

The issue wasn't immediately obvious - memory usage was slowly climbing during our stress tests, but only under specific connection patterns. After some investigation, I found that we weren't properly removing event listeners when WebSocket connections closed unexpectedly, leading to a classic memory leak scenario."`
    },
    technical: {
        title: "Technical Deep-Dive",
        content: `"Memory leak investigation in WebSocket connection pool implementation revealed critical event listener management oversight. Root cause analysis:

**Problem**: Event listeners attached via addEventListener() for 'close', 'error', and 'message' events were not being properly cleaned up when connections terminated unexpectedly (network timeouts, server restarts).

**Technical details**: JavaScript's garbage collector cannot reclaim closures that reference DOM elements or persistent objects when event listeners remain attached. Our connection pool maintained references to these listeners, preventing proper cleanup.

**Solution**: Implemented explicit removeEventListener() calls in our connection cleanup routine, reducing memory footprint by ~40% under high-churn scenarios."`
    },
    storytelling: {
        title: "Storytelling",
        content: `"The debugging session started like any other Tuesday afternoon. Our monitoring dashboards were showing a subtle but persistent increase in memory usage - the kind that makes you squint at graphs and wonder if you're imagining things.

But this wasn't imagination. Something was slowly eating away at our server's memory, one WebSocket connection at a time. The plot twist? It wasn't the connections themselves causing the leak - it was the ghosts they left behind.

Every time a connection closed unexpectedly (network hiccups, users closing laptops, the usual chaos of the internet), we were cleaning up the connection object but forgetting about the event listeners still clinging to memory. Like digital barnacles, they accumulated over time, each one a tiny reminder of a conversation that had ended.

The breakthrough moment came when I realized we needed to be more intentional about our goodbyes."`
    },
    minimalist: {
        title: "Minimalist",
        content: `**Issue**: WebSocket memory leak
**Cause**: Event listeners not cleaned up properly  
**Impact**: Memory usage climbing during stress tests
**Solution**: Added explicit removeEventListener() calls
**Result**: 40% memory reduction under high connection churn

**Key takeaway**: Always clean up event listeners when destroying objects."`
    },
    thought: {
        title: "Thought Leadership",
        content: `"This WebSocket memory leak investigation highlights a broader pattern in modern web development: the hidden costs of event-driven architecture. As we build increasingly connected applications, proper lifecycle management becomes not just a best practice, but a competitive advantage.

The industry's shift toward real-time features (live collaboration, instant messaging, real-time dashboards) means WebSocket implementations are becoming critical infrastructure. Yet many teams treat them as implementation details rather than architectural components that require careful design.

This experience reinforces why memory management discipline separates production-ready systems from prototype code. In an era where developers can deploy complex real-time features with just a few lines of code, understanding the underlying resource implications becomes increasingly valuable.

What this means for engineering leaders: invest in infrastructure knowledge, not just feature velocity."`
    }
};

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

// Initialize voice demo
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the animated title
    initAnimatedTitle();
    
    // Initialize the spectacular uroboro animation
    initUroboroAnimation();
    
    // Initialize unified carousel system
    initCarousels();
    
    // Initialize dynamic tickertape
    initTickertape();
    
    const voiceTabs = document.querySelectorAll('.voice-tab');
    const voiceContent = document.getElementById('voice-content');
    
    voiceTabs.forEach(tab => {
        tab.addEventListener('click', function() {
            // Remove active class from all tabs
            voiceTabs.forEach(t => t.classList.remove('active'));
            
            // Add active class to clicked tab
            this.classList.add('active');
            
            // Get voice type and update content
            const voiceType = this.getAttribute('data-voice');
            const example = voiceExamples[voiceType];
            
            if (example) {
                // Add fade effect
                voiceContent.style.opacity = '0.5';
                
                setTimeout(() => {
                    voiceContent.innerHTML = example.content;
                    voiceContent.style.opacity = '1';
                }, 150);
            }
        });
    });
    
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

// Unified Carousel System with backward compatibility
function initCarousels() {
    // Find all carousels on the page
    const carouselContainers = document.querySelectorAll('.carousel-container');
    
    carouselContainers.forEach(container => {
        // Handle both old and new button classes
        const navButtons = container.querySelectorAll('.carousel-nav-btn, .nav-btn');
        const slides = container.querySelectorAll('.carousel-slide, .feature-slide, .demo-slide');
        
        console.log('Initializing carousel:', {
            navButtons: navButtons.length,
            slides: slides.length,
            slideIds: Array.from(slides).map(s => s.id)
        });
        
        // Aggressively hide all slides and show first one
        slides.forEach((slide, index) => {
            slide.classList.remove('active');
            slide.style.display = 'none';
            slide.style.visibility = 'hidden';
            slide.style.opacity = '0';
            slide.style.position = 'absolute';
            slide.style.height = '0';
            slide.style.overflow = 'hidden';
            
            if (index === 0) {
                slide.classList.add('active');
                slide.style.display = 'block';
                slide.style.visibility = 'visible';
                slide.style.opacity = '1';
                slide.style.position = 'relative';
                slide.style.height = 'auto';
                slide.style.overflow = 'visible';
            }
        });
        
        // Set first nav button as active
        navButtons.forEach((btn, index) => {
            btn.classList.remove('active');
            if (index === 0) {
                btn.classList.add('active');
            }
        });
        
        // Add click handlers with compatibility for old and new data attributes
        navButtons.forEach(button => {
            button.addEventListener('click', function() {
                // Handle both new and old data attribute formats
                let targetId = this.getAttribute('data-carousel-target');
                
                // Fallback to old format
                if (!targetId) {
                    const feature = this.getAttribute('data-feature');
                    const demo = this.getAttribute('data-demo');
                    
                    if (feature) {
                        targetId = `${feature}-slide`;
                    } else if (demo) {
                        targetId = `${demo}-demo`;
                    }
                }
                
                console.log('Switching to:', targetId);
                
                // Remove active from all buttons and slides in this carousel
                navButtons.forEach(btn => btn.classList.remove('active'));
                
                // Aggressively hide all slides
                slides.forEach(slide => {
                    slide.classList.remove('active');
                    slide.style.display = 'none';
                    slide.style.visibility = 'hidden';
                    slide.style.opacity = '0';
                    slide.style.position = 'absolute';
                    slide.style.height = '0';
                    slide.style.overflow = 'hidden';
                });
                
                // Add active class to clicked button
                this.classList.add('active');
                
                // Show target slide
                const targetSlide = document.getElementById(targetId);
                if (targetSlide) {
                    targetSlide.classList.add('active');
                    targetSlide.style.display = 'block';
                    targetSlide.style.visibility = 'visible';
                    targetSlide.style.opacity = '1';
                    targetSlide.style.position = 'relative';
                    targetSlide.style.height = 'auto';
                    targetSlide.style.overflow = 'visible';
                    console.log('Activated slide:', targetSlide.id);
                } else {
                    console.error('Could not find slide:', targetId);
                }
            });
        });
    });
}

// Add some terminal-like typing effect to the hero demo (optional enhancement)
function typewriterEffect(element, text, speed = 50) {
    element.innerHTML = '';
    let i = 0;
    
    function typeChar() {
        if (i < text.length) {
            element.innerHTML += text.charAt(i);
            i++;
            setTimeout(typeChar, speed);
        }
    }
    
    typeChar();
}

// Optional: Add typing effect to demo command on page load
document.addEventListener('DOMContentLoaded', function() {
    const demoCommand = document.querySelector('.demo-command');
    if (demoCommand) {
        const originalText = demoCommand.textContent;
        // Small delay to let page settle
        setTimeout(() => {
            typewriterEffect(demoCommand, originalText, 30);
        }, 1000);
    }
});

// Dynamic Tickertape Generation - Simple Always-Visible Scroll 🐍
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
        bottomTickertape.title = 'Click to return to the beginning... 🐍';
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