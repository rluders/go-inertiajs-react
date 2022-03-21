require('./bootstrap')

import React from 'react'
import { render } from 'react-dom'
import {createInertiaApp, usePage} from '@inertiajs/inertia-react'
import { InertiaProgress } from '@inertiajs/progress'

createInertiaApp({
    title: (title) => {
        let appName = usePage().props.title
        return `${title} - ${appName}`
    },
    resolve: (name) => require(`./Pages/${name}`),
    setup({ el, App, props }) {
        return render(<App {...props} />, el)
    },
})

InertiaProgress.init({ color: '#4B5563' })
