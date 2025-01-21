document.addEventListener('DOMContentLoaded', () => {
  addFormValidationEventListener()
  addEventListenerForPasswordVisibilityToggle()
  addEventListenerForPasswordEditFields()
})

function addFormValidationEventListener() {
  'use strict'

  // Fetch all the forms we want to apply custom Bootstrap validation styles to
  var forms = document.querySelectorAll('.needs-validation')

  // Loop over them and prevent submission
  Array.prototype.slice.call(forms).forEach(function (form) {
    form.addEventListener(
      'submit',
      function (event) {
        if (!form.checkValidity()) {
          event.preventDefault()
          event.stopPropagation()
        }

        form.classList.add('was-validated')
      },
      false
    )
  })
}

function addEventListenerForPasswordVisibilityToggle() {
  const pwField = document.getElementById('togglePassword')
  if (pwField == null) {
    return
  }

  pwField.addEventListener('click', function () {
    const password = document.getElementById('password')
    const icon = this.querySelector('i')

    if (password.type === 'password') {
      password.type = 'text'
      icon.classList.remove('fa-eye')
      icon.classList.add('fa-eye-slash')
    } else {
      password.type = 'password'
      icon.classList.remove('fa-eye-slash')
      icon.classList.add('fa-eye')
    }
  })
}

function addEventListenerForPasswordEditFields() {
  const profileForm = document.getElementById('profileForm')

  if (profileForm == null) {
    return
  }
  profileForm.addEventListener('submit', function (e) {
    e.preventDefault()

    // Get all password fields
    const passwordFields = document.querySelectorAll('.password-field')
    const passwordValues = Array.from(passwordFields).map((field) =>
      field.value.trim()
    )

    // Check if any password field has a value
    const anyPasswordFilled = passwordValues.some((value) => value !== '')
    // Check if all password fields have values
    const allPasswordsFilled = passwordValues.every((value) => value !== '')

    // Reset validation state
    passwordFields.forEach((field) => {
      field.classList.remove('is-invalid')
    })

    // If any password field is filled, all must be filled
    if (anyPasswordFilled && !allPasswordsFilled) {
      passwordFields.forEach((field) => {
        if (!field.value.trim()) {
          field.classList.add('is-invalid')
        }
      })
      return
    }
  })
}
