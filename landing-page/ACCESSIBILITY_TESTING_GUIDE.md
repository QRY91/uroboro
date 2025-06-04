# Accessibility Testing Guide for uroboro Landing Page

## Quick Testing Gauntlet (15 minutes)

### 🎹 **Keyboard Navigation Test (3 minutes)**
1. **Tab through entire page** - Every interactive element should be reachable
2. **Press Tab once** - Skip link should appear and be clearly visible
3. **Press Enter on skip link** - Should jump to main content
4. **Tab to carousel** - Arrow keys should navigate between tabs
5. **Tab to CTA buttons** - Should have clear focus indicators

**Pass criteria:** ✅ All interactive elements reachable, clear focus indicators, no keyboard traps

### 👁️ **Screen Reader Test (5 minutes)**
**Using built-in screen readers:**
- **Windows:** NVDA (free) or Narrator
- **Mac:** VoiceOver (Cmd+F5)
- **Linux:** Orca

**Quick tests:**
1. **Skip link announces properly** when focused
2. **Headings read in order** (H1 → H2 → H3)
3. **Carousel tabs announce** "tab 1 of 3" etc.
4. **Images have descriptive alt text**
5. **Code blocks read as "code block"**

**Pass criteria:** ✅ Clear content structure, meaningful descriptions, no meaningless announcements

### 🎨 **Visual Accessibility Test (3 minutes)**
1. **Zoom to 200%** - Content should remain usable
2. **Check focus indicators** - Should be visible on all interactive elements
3. **Test in high contrast mode** (Windows: Alt+Shift+PrtScr)
4. **Disable animations** (Chrome DevTools → Rendering → Emulate CSS prefers-reduced-motion)

**Pass criteria:** ✅ Content scales properly, clear focus, works without animations

### 📱 **Mobile Touch Test (2 minutes)**
1. **Touch targets minimum 44px** - Easy to tap buttons
2. **Pinch zoom works** - Can zoom without breaking layout
3. **Orientation change** - Works in portrait/landscape

**Pass criteria:** ✅ Touch-friendly, responsive, orientation-independent

### 🔄 **Motion & Animation Test (2 minutes)**
1. **Set reduced motion preference:**
   - **Mac:** System Preferences → Accessibility → Display → Reduce motion
   - **Windows:** Settings → Ease of Access → Display → Show animations
   - **Browser:** Chrome DevTools → Rendering → Emulate CSS prefers-reduced-motion: reduce

2. **Verify animations disabled** - No spinning uroboro, no tickertape movement
3. **Test with motion enabled** - Smooth animations, no seizure triggers

**Pass criteria:** ✅ Respects motion preferences, no jarring movements

---

## Comprehensive Testing Toolkit

### 🛠️ **Automated Testing Tools**

#### **Browser Extensions (Install these first)**
```bash
# Chrome/Edge Extensions:
- axe DevTools (Free) - https://chrome.google.com/webstore/detail/axe-devtools-web-accessibility/lhdoppojpmngadmnindnejefpokejbdd
- WAVE Evaluation Tool - https://chrome.google.com/webstore/detail/wave-evaluation-tool/jbbplnpkjmmeebjpijfedlgcdilocofh
- Lighthouse (Built into Chrome DevTools)
- Accessibility Insights for Web - https://chrome.google.com/webstore/detail/accessibility-insights-fo/pbjjkligggfmakdaogkfomddhfmpjeni
```

#### **Command Line Tools**
```bash
# Install accessibility testing tools
npm install -g @axe-core/cli pa11y lighthouse

# Quick automated tests
axe http://localhost:3000/landing-page/
pa11y http://localhost:3000/landing-page/
lighthouse http://localhost:3000/landing-page/ --only-categories=accessibility
```

### 📋 **Manual Testing Checklist**

#### **1. Keyboard Navigation (WCAG 2.1.1, 2.1.2)**
- [ ] **Tab order is logical** (left-to-right, top-to-bottom)
- [ ] **Skip link works** (Tab → Enter → jumps to main content)
- [ ] **No keyboard traps** (can Tab out of all components)
- [ ] **Carousel navigation** (Arrow keys work in demo tabs)
- [ ] **Enter/Space activate buttons**
- [ ] **Escape closes modals** (if any)
- [ ] **Focus visible** on all interactive elements

#### **2. Screen Reader Testing (WCAG 4.1.2, 1.3.1)**
- [ ] **Page title descriptive** ("uroboro - The Self-Documenting Content Pipeline")
- [ ] **Headings hierarchical** (H1 → H2 → H3, no skipped levels)
- [ ] **Landmarks clear** (banner, main, navigation, contentinfo)
- [ ] **Lists announced properly** (value props, features, install steps)
- [ ] **Images have alt text** (especially demo GIFs)
- [ ] **Code blocks identified** as code regions
- [ ] **Carousel state announced** ("tab 1 of 3 selected")

#### **3. Visual Design (WCAG 1.4.3, 1.4.11)**
- [ ] **Color contrast sufficient** (4.5:1 for normal text, 3:1 for large)
- [ ] **Focus indicators visible** (2px orange outline)
- [ ] **Text scales to 200%** without horizontal scroll
- [ ] **Touch targets 44px minimum** on mobile
- [ ] **Content reflows properly** at different zoom levels

#### **4. Motion & Animation (WCAG 2.3.3, 2.2.2)**
- [ ] **Reduced motion respected** (animations disabled when requested)
- [ ] **No seizure triggers** (no flashing > 3 times per second)
- [ ] **Auto-animations can be paused** (hover/focus pauses carousel)
- [ ] **Essential motion only** (decorative animations can be disabled)

### 🔧 **Browser DevTools Testing**

#### **Chrome DevTools Accessibility Panel**
1. **Open DevTools** (F12)
2. **Go to Lighthouse tab** → Run accessibility audit
3. **Elements panel** → Accessibility tree view
4. **Rendering panel** → Emulate vision deficiencies
5. **Console** → Check for accessibility warnings

#### **Testing Commands in Console**
```javascript
// Check for missing alt text
document.querySelectorAll('img:not([alt])').length

// Check for empty headings
document.querySelectorAll('h1:empty, h2:empty, h3:empty, h4:empty, h5:empty, h6:empty').length

// Check for buttons without accessible names
document.querySelectorAll('button:not([aria-label]):not([aria-labelledby])').length

// Check focus order
document.querySelectorAll('[tabindex]')
```

### 🎯 **Specific uroboro Tests**

#### **Skip Link Test**
```bash
1. Load page
2. Press Tab once
3. ✅ Skip link appears below ticker
4. Press Enter
5. ✅ Focus moves to "main-content"
6. ✅ Visual focus indicator visible
```

#### **Carousel Accessibility Test**
```bash
1. Tab to demo carousel
2. ✅ First tab focused and announced as "selected"
3. Press Right Arrow
4. ✅ Next tab selected, content changes
5. ✅ Screen reader announces tab change
6. Press Home
7. ✅ First tab selected
8. Press End
9. ✅ Last tab selected
```

#### **Demo GIF Alt Text Test**
```bash
Screen reader should announce:
"Animated demonstration showing core uroboro workflow: 
capture command followed by status command followed by 
publish command, displaying terminal output for each step"
```

### 📊 **Automated Testing Script**

Create `test-accessibility.js`:
```javascript
const { chromium } = require('playwright');
const AxeBuilder = require('@axe-core/playwright').default;

async function testAccessibility() {
    const browser = await chromium.launch();
    const page = await browser.newPage();
    
    await page.goto('http://localhost:3000/landing-page/');
    
    const accessibilityScanResults = await new AxeBuilder({ page }).analyze();
    
    console.log('Accessibility violations:', accessibilityScanResults.violations.length);
    
    if (accessibilityScanResults.violations.length > 0) {
        console.log('Violations found:');
        accessibilityScanResults.violations.forEach(violation => {
            console.log(`- ${violation.id}: ${violation.description}`);
        });
    } else {
        console.log('✅ No accessibility violations found!');
    }
    
    await browser.close();
}

testAccessibility();
```

### 🚀 **Quick Start Testing**

**1. Install browser extension:**
```bash
Install axe DevTools extension for your browser
```

**2. Run quick scan:**
```bash
# Open landing page
# F12 → axe DevTools → Scan all of my page
# Fix any violations found
```

**3. Manual keyboard test:**
```bash
# Close all menus/modals
# Click in address bar
# Press Tab repeatedly through entire page
# Verify every interactive element is reachable
# Verify focus indicators are visible
```

**4. Screen reader test:**
```bash
# Windows: Windows+Ctrl+Enter (Narrator)
# Mac: Cmd+F5 (VoiceOver)  
# Navigate with arrow keys and Tab
# Verify content makes sense when read aloud
```

### ✅ **Success Criteria**

**Perfect Score Requirements:**
- **0 axe violations**
- **100% keyboard accessible**
- **Lighthouse accessibility score: 100**
- **Works with screen readers**
- **Respects motion preferences**
- **Color contrast ≥ 4.5:1**
- **Touch targets ≥ 44px**

### 🏆 **Testing Schedule**

**During Development:**
- **Every commit:** Quick keyboard test (1 min)
- **Every feature:** Screen reader test (2 min)
- **Before push:** Full gauntlet (15 min)

**Before Release:**
- **Full automated scan**
- **Multi-browser testing**
- **Real assistive technology testing**
- **Mobile device testing**

This testing approach ensures uroboro's landing page is truly accessible to everyone! 🌟 

---

*"Microblogging for people too busy to microblog"* - the uroboro philosophy 