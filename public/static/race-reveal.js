(function () {
  'use strict';

  var START_DELAY_MS  = 900;   // pause before the first race loads in
  var SCAN_PHASE_MS   = 550;   // "SIMULATING" scanning bar duration
  var RESULT_PHASE_MS = 950;   // result badge display before moving on

  var cells = document.querySelectorAll('[data-race-cell]');
  if (!cells.length) return;

  var total = cells.length;

  // Score strip elements
  var counterEl = document.getElementById('race-counter');
  var winsEl    = document.getElementById('live-wins');
  var lossesEl  = document.getElementById('live-losses');
  var scoreHero = document.getElementById('score-hero');
  var tierEl    = document.getElementById('tier-display');
  var postRace  = document.getElementById('post-race');

  // Spotlight elements
  var spotlight    = document.getElementById('race-spotlight');
  var spotLabel    = document.getElementById('spot-race-label');
  var spotIcon     = document.getElementById('spot-circuit-icon');
  var spotName     = document.getElementById('spot-name');
  var spotScanning = document.getElementById('spot-scanning');
  var spotBadge    = document.getElementById('spot-result-badge');

  var ICONS = { street: '🏙️', technical: '🔧', wet: '🌧️', normal: '🏁' };

  var wins = 0, losses = 0;

  function circuitIcon(type) {
    return ICONS[type] || ICONS.normal;
  }

  function enterScanState(raceNum, raceName, circuitType) {
    spotLabel.textContent = 'RACE ' + raceNum + ' / ' + total;
    spotIcon.textContent  = circuitIcon(circuitType);
    spotName.textContent  = raceName || '—';

    // Hide result badge, show scan bar
    spotBadge.style.display   = 'none';
    spotBadge.className       = '';
    spotScanning.style.display = 'flex';

    if (counterEl) counterEl.textContent = 'RACE ' + raceNum + ' / ' + total;
  }

  function enterResultState(won, dnf) {
    spotScanning.style.display = 'none';

    var cls  = won ? 'win' : (dnf ? 'dnf' : 'loss');
    var text = won ? 'WIN'  : (dnf ? 'DNF' : 'LOSS');

    // Remove animation class, force reflow, re-add — makes it fire every time
    spotBadge.className       = '';
    spotBadge.textContent     = text;
    spotBadge.style.display   = 'block';
    spotBadge.getBoundingClientRect();
    spotBadge.className = cls;
  }

  function revealCalendarEntry(cell, won, dnf) {
    var dot = cell.querySelector('.race-dot');
    if (dot) {
      if (won)       dot.classList.add('win');
      else if (dnf)  dot.classList.add('dnf');
      else           dot.classList.add('loss');
    }
    cell.classList.add('revealed');
  }

  function updateScore(won, dnf) {
    if (won)       wins++;
    else if (!dnf) losses++;

    if (winsEl)   winsEl.textContent   = wins;
    if (lossesEl) lossesEl.textContent = losses;

    if (scoreHero) {
      scoreHero.style.transition = 'transform .15s cubic-bezier(.34,1.56,.64,1)';
      scoreHero.style.transform  = 'scale(1.08)';
      setTimeout(function () { scoreHero.style.transform = 'scale(1)'; }, 160);
    }
  }

  function runRace(idx) {
    if (idx >= total) {
      onAllRevealed();
      return;
    }

    var cell        = cells[idx];
    var won         = cell.dataset.won         === 'true';
    var dnf         = cell.dataset.dnf         === 'true';
    var raceName    = cell.dataset.raceName    || '—';
    var circuitType = cell.dataset.circuitType || 'normal';

    // Phase 1: enter scanning state
    enterScanState(idx + 1, raceName, circuitType);

    // Phase 2: show result, reveal calendar entry, update score
    setTimeout(function () {
      enterResultState(won, dnf);
      revealCalendarEntry(cell, won, dnf);
      updateScore(won, dnf);

      // Phase 3: move to next race
      setTimeout(function () {
        runRace(idx + 1);
      }, RESULT_PHASE_MS);

    }, SCAN_PHASE_MS);
  }

  function onAllRevealed() {
    // Transition spotlight to "complete" state
    if (spotlight) spotlight.classList.add('complete');
    spotScanning.style.display = 'none';
    spotName.textContent  = wins + ' — ' + losses;
    spotBadge.className   = 'complete';
    spotBadge.textContent = 'SEASON COMPLETE';
    spotBadge.style.display = 'block';
    spotBadge.getBoundingClientRect();
    spotBadge.className   = 'complete';

    if (counterEl) counterEl.textContent = 'SEASON COMPLETE';

    // Fade in tier
    if (tierEl) {
      tierEl.style.opacity   = '1';
      tierEl.style.transform = 'translateY(0)';
    }

    // Fade in post-race section
    setTimeout(function () {
      if (postRace) {
        postRace.style.display  = 'block';
        postRace.style.opacity  = '0';
        postRace.style.transition = 'opacity .5s ease';
        postRace.getBoundingClientRect();
        postRace.style.opacity  = '1';
      }
    }, 400);
  }

  setTimeout(function () { runRace(0); }, START_DELAY_MS);
})();
