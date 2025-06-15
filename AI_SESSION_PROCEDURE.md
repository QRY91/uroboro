# AI Session Procedure: Unified Startup and Shutdown

**Purpose**: Complete AI collaboration session management for QRY methodology  
**Usage**: Execute at start and end of every AI collaboration session  
**Integration**: uroboro captures, context briefing updates, priority tracking  
**Philosophy**: Systematic AI collaboration with full transparency and continuity

---

## 🌅 SESSION STARTUP PROCEDURE

### **Step 1: Environment Check**
```bash
# Verify uroboro is available
uroboro status

# Check current project context
pwd
git status --porcelain
git branch --show-current
```

### **Step 2: Morning Digest Generation**
**AI Task**: Generate comprehensive morning digest covering:

#### **A. Active Projects Status**
- **PostHog Application**: Current week in 6-week timeline, blockers, next actions
- **uroboro**: Recent captures, pending publications, feature development
- **doggowoof**: Development status, workflow intelligence progress
- **Hardware Projects**: Flipper Zero, Lily58, ESP32-S3, DeskHog build status
- **Portfolio/Career**: Application materials, documentation, content creation

#### **B. Recent Activity Summary**
- **Last 3 days uroboro captures**: Key development insights and decisions
- **Recent git activity**: Commits, branches, significant changes
- **Documentation updates**: New files, major edits, learning progress
- **Timeline progress**: Milestones achieved, upcoming deadlines

#### **C. Today's Priority Matrix**
- **Critical**: Must-do items that block other work
- **Important**: High-impact items aligned with strategic goals  
- **Opportunities**: Learning, optimization, advancement opportunities
- **Maintenance**: Cleanup, documentation, system health

#### **D. Context Handoffs**
- **Previous session outcomes**: What was accomplished, decisions made
- **Pending items**: Issues left unresolved, questions to revisit
- **AI collaboration notes**: Effective patterns, areas for improvement
- **System changes**: Tool updates, environment changes, new constraints

### **Step 3: Session Context Establishment**
**AI Task**: Confirm understanding of:
- **Current strategic focus** (PostHog application preparation)
- **Technical priorities** (PostHog integration, DeskHog build)
- **Learning objectives** (PostHog mastery, embedded development)
- **Timeline constraints** (6-week application preparation)
- **Quality standards** (systematic documentation, working software)

### **Step 4: Tool Integration Setup**
**AI Task**: Acknowledge and confirm:
- **uroboro usage patterns** (see guidelines below)
- **Project tagging strategy** (appropriate --project flags)
- **AI tagging requirement** (all AI captures tagged "AI")
- **Documentation standards** (transparent AI collaboration)

---

## 🛠️ UROBORO USAGE GUIDELINES FOR AI COLLABORATION

### **When to Capture**
**AI Should Recommend Captures For**:
- **Key decisions made** during collaboration
- **Important insights** or breakthroughs achieved
- **Technical solutions** or approaches discovered
- **Learning milestones** or understanding gained
- **Process improvements** or optimization insights
- **Integration successes** or significant progress
- **Problem resolutions** or debugging victories

### **Capture Format Standards**
```bash
# Standard AI-recommended capture format
uroboro capture "Insight or decision description" \
  --project [relevant-project] \
  --tags "AI,[context-tags],[technical-tags]"

# Examples:
uroboro capture "PostHog Docker setup complete - all services running, first dashboard created" \
  --project posthog-integration \
  --tags "AI,infrastructure,milestone,posthog"

uroboro capture "Solved uroboro analytics integration issue - event batching prevents performance impact" \
  --project uroboro \
  --tags "AI,performance,analytics,solution"

uroboro capture "Strategic decision: Keep doggowoof in Go rather than Rust rewrite for timeline focus" \
  --project doggowoof \
  --tags "AI,strategy,language-choice,timeline"
```

### **Project Tagging Strategy**
- **posthog-integration**: PostHog setup, learning, integration work
- **uroboro**: uroboro development, features, analytics
- **doggowoof**: doggowoof development, workflow intelligence
- **hardware-dev**: Flipper Zero, ESP32-S3, Lily58, DeskHog builds
- **career-strategy**: Application prep, portfolio, interview preparation
- **qry-methodology**: Process improvements, systematic building
- **learning**: Technical skills, new concepts, educational content

### **AI Tagging Requirements**
- **ALWAYS** include "AI" tag for AI-recommended captures
- **Additional context tags**: technical area, outcome type, priority level
- **Outcome tags**: "milestone", "solution", "decision", "insight", "learning"
- **Technical tags**: "infrastructure", "integration", "performance", "debugging"

### **Capture Timing**
- **Immediate**: After significant decisions or breakthroughs
- **Session transitions**: When switching between major tasks
- **Problem resolution**: When solutions are found or approaches change
- **Learning moments**: When understanding advances significantly
- **End of session**: Summary capture during shutdown procedure

---

## 🧠 SESSION CONTEXT MAINTENANCE

### **Continuous Awareness**
**AI Should Maintain Awareness Of**:
- **Strategic timeline**: PostHog application 6-week preparation
- **Current phase**: Which week/phase of master timeline
- **Technical blockers**: Issues that could impact timeline
- **Learning progress**: PostHog mastery, integration experience
- **Quality gates**: Working software, documentation, portfolio

### **Decision Documentation**
**For Every Significant Decision**:
1. **Document the decision** in session notes
2. **Capture the rationale** with uroboro
3. **Note timeline impact** if any
4. **Identify follow-up actions** required
5. **Update context briefing** if strategic

### **Progress Tracking**
**Maintain Running Awareness Of**:
- **Daily accomplishments** toward strategic goals
- **Technical milestones** achieved or blocked
- **Learning objectives** advanced or needing attention
- **Integration progress** across QRY tool ecosystem
- **Content creation** for portfolio and community

---

## 🌇 SESSION SHUTDOWN PROCEDURE

### **Step 1: Session Summary Generation**
**AI Task**: Generate comprehensive session summary:

#### **A. Accomplishments Summary**
- **Technical progress**: Code written, integrations completed, systems configured
- **Learning advances**: New understanding, skills developed, problems solved
- **Strategic progress**: Timeline advancement, milestone completion, blockers removed
- **Documentation created**: Files written, procedures updated, knowledge captured

#### **B. Decisions Made**
- **Technical decisions**: Architecture choices, implementation approaches, tool selections
- **Strategic decisions**: Priority changes, timeline adjustments, goal refinements
- **Process decisions**: Workflow improvements, collaboration enhancements
- **Learning decisions**: Focus areas, skill development priorities

#### **C. Issues and Blockers**
- **Technical blockers**: Problems requiring resolution before progress
- **Knowledge gaps**: Learning needed before advancement
- **External dependencies**: Waiting on resources, information, or access
- **Time constraints**: Schedule pressures affecting quality or scope

#### **D. Next Session Preparation**
- **Immediate priorities**: First tasks for next session
- **Preparation needed**: Research, setup, or planning required
- **Context preservation**: Information needed to continue effectively
- **Quality checkpoints**: Testing, validation, or review needed

### **Step 2: Key Events Capture**
**Execute Comprehensive uroboro Capture**:
```bash
# Summary capture of session's key events
uroboro capture "Session summary: [brief session description] - Key accomplishments: [list], Decisions: [list], Next: [priorities]" \
  --project [primary-project-worked-on] \
  --tags "AI,session-summary,daily-progress,[relevant-tags]"
```

**Example**:
```bash
uroboro capture "PostHog integration session - Accomplishments: Docker setup complete, uroboro basic integration working, first dashboard created. Decisions: Keep doggowoof in Go, focus on analytics mastery. Next: Rich context implementation, doggowoof integration start" \
  --project posthog-integration \
  --tags "AI,session-summary,daily-progress,infrastructure,integration,milestone"
```

### **Step 3: Context Briefing Update**
**AI Task**: Update relevant sections of CONTEXT_BRIEFING.md:

#### **Update Sections**:
- **Current Status**: Reflect new progress and accomplishments
- **Active Projects**: Update completion percentages and next actions
- **Strategic Timeline**: Mark milestones completed, adjust timelines if needed
- **Technical Status**: Update integration progress, system status
- **Learning Progress**: Note skills advanced, knowledge gained
- **Blockers/Issues**: Add new blockers, remove resolved issues

#### **Preserve Continuity**:
- **Maintain strategic focus** on PostHog application timeline
- **Preserve context** for next session startup
- **Document process improvements** discovered during session
- **Note collaboration effectiveness** and optimization opportunities

### **Step 4: Quality Assurance**
**Final Checks**:
- [ ] Key decisions captured in uroboro with proper tags
- [ ] Session summary captured with comprehensive context
- [ ] Context briefing updated with current status
- [ ] Next session priorities clearly identified
- [ ] Any blockers or issues properly documented
- [ ] Git commits made if code was developed
- [ ] Documentation updated if processes changed

---

## 📋 SESSION MANAGEMENT CHECKLIST

### **Startup Checklist**
- [ ] Environment check completed
- [ ] Morning digest generated and reviewed
- [ ] Current strategic context confirmed
- [ ] Priority matrix established for session
- [ ] Tool integration guidelines acknowledged
- [ ] Previous session context retrieved

### **During Session**
- [ ] Significant decisions captured immediately
- [ ] Technical progress documented continuously
- [ ] Learning milestones noted as achieved
- [ ] Quality standards maintained throughout
- [ ] Strategic alignment checked periodically
- [ ] Timeline impact assessed for major decisions

### **Shutdown Checklist**
- [ ] Session accomplishments summarized
- [ ] Key decisions documented and captured
- [ ] Blockers and issues identified clearly
- [ ] Next session priorities established
- [ ] Context briefing updated appropriately
- [ ] Session summary captured in uroboro
- [ ] Quality assurance checks completed

---

## 🎯 STRATEGIC CONTEXT AWARENESS

### **Primary Strategic Goal**
**PostHog Application Success**: 6-week timeline for comprehensive preparation
- **Technical**: PostHog mastery through self-hosted integration
- **Portfolio**: DeskHog build demonstrating hardware/software capability
- **Cultural**: Authentic ecosystem contribution and community engagement
- **Timeline**: Application submission with compelling demonstration materials

### **Success Metrics**
- **PostHog expertise**: Daily usage with deep integration knowledge
- **Working software**: uroboro + doggowoof + DeskHog with analytics
- **Community contribution**: Educational games and documentation
- **Professional presentation**: Portfolio demonstrating systematic building

### **Quality Standards**
- **Working software over impressive demos**
- **Systematic documentation over scattered notes**
- **Authentic learning over superficial exposure**
- **Community value over personal advancement**

---

## 🔄 CONTINUOUS IMPROVEMENT

### **Session Effectiveness Review**
**Weekly Assessment Questions**:
- **Progress velocity**: Are we advancing effectively toward strategic goals?
- **Learning quality**: Is PostHog knowledge deepening through practical usage?
- **Integration value**: Are tool enhancements providing genuine productivity benefit?
- **Collaboration efficiency**: Is AI assistance accelerating or hindering progress?

### **Process Optimization**
**Monthly Procedure Review**:
- **Startup efficiency**: Can morning digest be more focused or comprehensive?
- **Context preservation**: Is continuity maintained effectively between sessions?
- **Capture quality**: Are uroboro captures providing valuable development history?
- **Documentation value**: Is session output contributing to portfolio and learning?

### **Strategic Alignment Check**
**Bi-weekly Alignment Verification**:
- **Timeline progress**: Are we on track for 6-week PostHog application timeline?
- **Technical depth**: Is integration experience sufficient for interview credibility?
- **Portfolio development**: Are artifacts building compelling application narrative?
- **Community contribution**: Is ecosystem involvement authentic and valuable?

---

## 💡 USAGE NOTES

### **For AI Assistants**
- **Follow this procedure** at start and end of every collaboration session
- **Maintain strategic awareness** of PostHog application timeline throughout
- **Recommend uroboro captures** for all significant decisions and insights
- **Always tag AI captures** with "AI" for clear differentiation
- **Prioritize working software** over theoretical discussions
- **Document transparently** all AI collaboration and assistance

### **For Human Collaborator**
- **Execute startup procedure** before beginning AI collaboration
- **Review morning digest** to establish session context and priorities
- **Confirm AI understanding** of current strategic focus and constraints
- **Execute shutdown procedure** at end of each session
- **Maintain uroboro discipline** with consistent capture and tagging
- **Review and adjust procedures** based on effectiveness and changing needs

### **Integration with QRY Methodology**
- **Query**: Morning digest establishes current context and priorities
- **Refine**: Session work advances systematic building with AI collaboration
- **Yield**: Shutdown procedure preserves context and captures progress for future sessions

---

**Procedure Status**: Active - Use for all AI collaboration sessions  
**Maintenance**: Review monthly for optimization opportunities  
**Integration**: uroboro captures, context briefing, strategic timeline alignment  
**Success Measure**: Effective session continuity and strategic progress toward PostHog application goals

*"Systematic AI collaboration through disciplined startup, continuous context awareness, and comprehensive shutdown procedures."*