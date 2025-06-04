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

// Unified Carousel System with Full Accessibility Support
function initCarousels() {
    // Find all carousel containers
    const carousels = document.querySelectorAll('.carousel-container');
    
    carousels.forEach(carousel => {
        const navButtons = carousel.querySelectorAll('.carousel-nav-btn, .nav-btn');
        const slides = carousel.querySelectorAll('.carousel-slide, .feature-slide, .demo-slide');
        let currentSlide = 0;
        
        // Function to show specific slide with accessibility support
        function showSlide(index, fromKeyboard = false) {
            // Hide all slides
            slides.forEach((slide, i) => {
                if (i === index) {
                    slide.classList.add('active');
                    slide.setAttribute('aria-hidden', 'false');
                    slide.setAttribute('tabindex', '0');
                } else {
                    slide.classList.remove('active');
                    slide.setAttribute('aria-hidden', 'true');
                    slide.setAttribute('tabindex', '-1');
                }
            });
            
            // Update nav button states
            navButtons.forEach((btn, i) => {
                if (i === index) {
                    btn.classList.add('active');
                    btn.setAttribute('aria-selected', 'true');
                    btn.setAttribute('tabindex', '0');
                    if (fromKeyboard) {
                        btn.focus();
                    }
                } else {
                    btn.classList.remove('active');
                    btn.setAttribute('aria-selected', 'false');
                    btn.setAttribute('tabindex', '-1');
                }
            });
            
            currentSlide = index;
            
            // Announce slide change to screen readers
            if (fromKeyboard) {
                const slideName = navButtons[index]?.textContent || `Slide ${index + 1}`;
                announceToScreenReader(`Showing ${slideName}`);
            }
        }
        
        // Add click and keyboard event listeners to nav buttons
        navButtons.forEach((btn, index) => {
            btn.addEventListener('click', (e) => {
                e.preventDefault();
                showSlide(index);
            });
            
            // Enhanced keyboard navigation
            btn.addEventListener('keydown', (e) => {
                switch(e.key) {
                    case 'ArrowLeft':
                        e.preventDefault();
                        const prevIndex = currentSlide > 0 ? currentSlide - 1 : navButtons.length - 1;
                        showSlide(prevIndex, true);
                        break;
                    case 'ArrowRight':
                        e.preventDefault();
                        const nextIndex = currentSlide < navButtons.length - 1 ? currentSlide + 1 : 0;
                        showSlide(nextIndex, true);
                        break;
                    case 'Home':
                        e.preventDefault();
                        showSlide(0, true);
                        break;
                    case 'End':
                        e.preventDefault();
                        showSlide(navButtons.length - 1, true);
                        break;
                    case 'Enter':
                    case ' ':
                        e.preventDefault();
                        showSlide(index, true);
                        break;
                }
            });
        });
        
        // Initialize first slide
        showSlide(0);
        
        // Auto-advance carousel if requested (with pause on focus/hover)
        const autoAdvance = carousel.getAttribute('data-auto-advance');
        if (autoAdvance) {
            let autoAdvanceInterval;
            const intervalTime = parseInt(autoAdvance) || 5000;
            
            function startAutoAdvance() {
                autoAdvanceInterval = setInterval(() => {
                    if (document.visibilityState === 'visible') {
                        const nextIndex = currentSlide < navButtons.length - 1 ? currentSlide + 1 : 0;
                        showSlide(nextIndex);
                    }
                }, intervalTime);
            }
            
            function stopAutoAdvance() {
                clearInterval(autoAdvanceInterval);
            }
            
            // Start auto-advance
            startAutoAdvance();
            
            // Pause on hover or focus (accessibility requirement)
            carousel.addEventListener('mouseenter', stopAutoAdvance);
            carousel.addEventListener('mouseleave', startAutoAdvance);
            carousel.addEventListener('focusin', stopAutoAdvance);
            carousel.addEventListener('focusout', startAutoAdvance);
            
            // Pause when page is hidden
            document.addEventListener('visibilitychange', () => {
                if (document.visibilityState === 'hidden') {
                    stopAutoAdvance();
                } else {
                    startAutoAdvance();
                }
            });
        }
    });
}

// Screen reader announcement function
function announceToScreenReader(message) {
    const announcement = document.createElement('div');
    announcement.setAttribute('aria-live', 'polite');
    announcement.setAttribute('aria-atomic', 'true');
    announcement.className = 'sr-only';
    announcement.textContent = message;
    
    document.body.appendChild(announcement);
    
    // Remove after announcement
    setTimeout(() => {
        document.body.removeChild(announcement);
    }, 1000);
}

// Enhanced reduced motion detection
function respectsReducedMotion() {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

// Initialize voice demo with accessibility improvements
function initVoiceDemo() {
    const voiceTabs = document.querySelectorAll('.voice-tab');
    const voiceContent = document.getElementById('voice-content');
    
    voiceTabs.forEach((tab, index) => {
        // Add ARIA attributes
        tab.setAttribute('role', 'tab');
        tab.setAttribute('aria-selected', index === 0 ? 'true' : 'false');
        tab.setAttribute('tabindex', index === 0 ? '0' : '-1');
        
        tab.addEventListener('click', function() {
            // Update all tabs
            voiceTabs.forEach((t, i) => {
                const isSelected = t === this;
                t.classList.toggle('active', isSelected);
                t.setAttribute('aria-selected', isSelected);
                t.setAttribute('tabindex', isSelected ? '0' : '-1');
            });
            
            // Get voice type and update content
            const voiceType = this.getAttribute('data-voice');
            const example = voiceExamples[voiceType];
            
            if (example && voiceContent) {
                voiceContent.innerHTML = `
                    <div class="voice-input">
                        <h4>Input:</h4>
                        <p>"Fixed memory leak in WebSocket connections - cut memory usage by 40%"</p>
                    </div>
                    <div class="voice-output">
                        <h4>Generated ${example.title}:</h4>
                        <p>${example.content}</p>
                    </div>
                `;
                
                // Announce change to screen readers
                announceToScreenReader(`Showing ${example.title} writing style`);
            }
        });
        
        // Keyboard navigation for voice tabs
        tab.addEventListener('keydown', (e) => {
            switch(e.key) {
                case 'ArrowLeft':
                    e.preventDefault();
                    const prevTab = index > 0 ? voiceTabs[index - 1] : voiceTabs[voiceTabs.length - 1];
                    prevTab.click();
                    prevTab.focus();
                    break;
                case 'ArrowRight':
                    e.preventDefault();
                    const nextTab = index < voiceTabs.length - 1 ? voiceTabs[index + 1] : voiceTabs[0];
                    nextTab.click();
                    nextTab.focus();
                    break;
            }
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