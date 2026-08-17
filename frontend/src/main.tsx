import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import {I18nProvider} from './i18n'
import {ThemeProvider} from './theme'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <ThemeProvider>
            <I18nProvider>
                <App/>
            </I18nProvider>
        </ThemeProvider>
    </React.StrictMode>
)
