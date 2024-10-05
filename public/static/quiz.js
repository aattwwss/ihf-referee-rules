const quizID = window.location.pathname.split('/').reverse()[0];
const QUIZ_CHOICE_MAP_KEY = `${quizID}-ChoiceMap`;

// Load state from localStorage
function loadState() {
    const choiceCheckboxes = document.querySelectorAll('.choice input[type="checkbox"]');
    const choiceCheckMap = JSON.parse(localStorage.getItem(QUIZ_CHOICE_MAP_KEY)) || {};
    choiceCheckboxes.forEach(checkbox => {
        const isChecked = choiceCheckMap[`${checkbox.name}-${checkbox.value}`];
        if (isChecked) {
            checkbox.checked = true;
        }
    });
}

// Save state to localStorage
function saveState() {
    const choiceCheckboxes = document.querySelectorAll('.choice input[type="checkbox"]');
    const choiceCheckMap = {};
    choiceCheckboxes.forEach(checkbox => {
        choiceCheckMap[`${checkbox.name}-${checkbox.value}`] = checkbox.checked;
    });
    localStorage.setItem(QUIZ_CHOICE_MAP_KEY, JSON.stringify(choiceCheckMap));
}

document.addEventListener('DOMContentLoaded', loadState);

document.querySelectorAll('.choice input[type="checkbox"]').forEach(checkbox => {
    checkbox.addEventListener('change', saveState);
});

// onclick quiz-submit-button, clear state
document.getElementById('quiz-submit-button').addEventListener('click', function () {
    // localStorage.setItem(QUIZ_CHOICE_MAP_KEY, JSON.stringify({}));
    localStorage.removeItem(QUIZ_CHOICE_MAP_KEY);
});
