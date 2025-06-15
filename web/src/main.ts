import './app.css'
import App from './App.svelte'

// Initialize the Svelte app
const app = new App({
  target: document.getElementById('app')!,
})

// Hot Module Replacement (HMR) support for development
if (import.meta.hot) {
  import.meta.hot.accept()
}

export default app
