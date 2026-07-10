package i18n

// tableES is the Spanish string table (design D1). Registered into the
// package-level tables map via init(). MUST carry the identical key set as
// tableEN (enforced by TestTableKeyParity in parity_test.go).
var tableES = map[string]string{
	// windows.install.* — install stream literals.
	"windows.install.starting":      "Iniciando instalación.",
	"windows.install.finished_ok":   "Instalación finalizada con éxito.",
	"windows.install.finished_fail": "La instalación falló.",

	// windows.uninstall.* — uninstall stream literals.
	"windows.uninstall.starting":      "Iniciando desinstalación.",
	"windows.uninstall.finished_ok":   "Desinstalación finalizada con éxito.",
	"windows.uninstall.finished_fail": "La desinstalación falló.",
	"windows.uninstall.plan_failed":   "Falló el plan de desinstalación.",

	// windows.phase.* — install phase messages.
	"windows.phase.prepare":  "Preparando instalación.",
	"windows.phase.apply":    "Aplicando pasos de instalación.",
	"windows.phase.rollback": "Revirtiendo cambios.",
	"windows.phase.default":  "Ejecutando instalación.",

	// windows.uninstall.phase.* — uninstall phase messages.
	"windows.uninstall.phase.prepare":  "Preparando desinstalación.",
	"windows.uninstall.phase.apply":    "Eliminando elementos instalados.",
	"windows.uninstall.phase.rollback": "Restaurando el estado anterior.",
	"windows.uninstall.phase.default":  "Ejecutando desinstalación.",

	// windows.step.* — per-step status messages.
	"windows.step.started":     "Paso iniciado.",
	"windows.step.succeeded":   "Paso completado.",
	"windows.step.failed":      "Paso fallido.",
	"windows.step.rolled_back": "Paso de reversión completado.",
	"windows.step.degraded":    "Paso completado con advertencias.",

	// windows.dryrun.* — dry-run step messages.
	"windows.dryrun.planned": "Paso de prueba (dry-run) planificado.",
	"windows.dryrun.listed":  "Paso de prueba (dry-run) listado.",

	// component.fallback.* — component-description fallbacks.
	"component.fallback.security":    "Protección básica para una configuración más segura. Esto siempre se instala.",
	"component.fallback.external":    "Descarga y configura una herramienta externa.",
	"component.fallback.config":      "Aplica la configuración recomendada.",
	"component.fallback.skill":       "Agrega ayudantes de flujo de trabajo guiado.",
	"component.fallback.generic_fmt": "Instala %s.",

	// mode.{lite,full,custom}.{label,desc,long} — install-mode options.
	"mode.lite.label": "Rápido",
	"mode.lite.desc":  "Instalación rápida para empezar a trabajar de inmediato.",
	"mode.lite.long":  "El modo Rápido instala solo el sustrato mínimo útil para arrancar de inmediato. Configura los harnesses esenciales—OpenSpec, Engram, Context7, el orquestador SDD, los permisos seguros y el comando de starters—y omite las skills de fundación y las pantallas de configuración avanzada. Elegilo cuando querés empezar ya y no necesitás la fundación guiada completa; siempre podés sumar el resto después con el modo Completo o Personalizado.",

	"mode.full.label": "Completo",
	"mode.full.desc":  "Instalación completa recomendada con todas las herramientas clave.",
	"mode.full.long":  "El modo Completo instala todo el kit recomendado: el sustrato de Rápido más las skills de fundación (active-orchestrator, kb-creator, roadmap-generator, agent-instruction, skill-registry) y las de terceros (find-skill, skill-creator). Deja el proyecto listo para el ciclo OPSX completo con la fundación guiada incluida. Tarda un poco más que Rápido pero no deja afuera nada importante. Elegilo cuando querés todo desde el primer día y no ir sumando piezas a mano.",

	"mode.custom.label": "Personalizado",
	"mode.custom.desc":  "Elegí exactamente qué instalar.",
	"mode.custom.long":  "El modo Personalizado te deja elegir exactamente qué harnesses instalar, uno por uno. No se instala nada fuera de tu selección, salvo el piso de seguridad obligatorio (los permisos), que siempre entra. Usalo cuando ya conocés tu stack y querés control total sobre qué queda y qué no. Si preferís una decisión ya tomada, Rápido o Completo son más directos.",

	// tier.{estricto,balanceado,bypass}.{label,desc,long} + tier.bypass.warning.
	"tier.estricto.label": "Estricto",
	"tier.estricto.desc":  "El agente debe pedir permiso para cada operación. Máxima fricción, máxima seguridad.",
	"tier.estricto.long":  "Estricto configura los permisos del agente para que pida confirmación antes de cada operación. Aplica el conjunto de reglas más restrictivo sobre el piso de seguridad: nada corre sin tu visto bueno. Es el nivel de mayor fricción y mayor seguridad. Elegilo para proyectos sensibles o si recién empezás y querés revisar todo antes de que pase; si te traba el ritmo, Balanceado afloja lo repetitivo.",

	"tier.balanceado.label": "Balanceado",
	"tier.balanceado.desc":  "Lista de permitidos curada para operaciones seguras y repetitivas. Punto de partida recomendado.",
	"tier.balanceado.long":  "Balanceado configura una lista de permitidos curada para que las operaciones seguras y repetitivas corran sin interrupciones, mientras las acciones más riesgosas siguen pidiendo confirmación. Es el punto de equilibrio entre fricción y autonomía sobre el piso de seguridad. Es el punto de partida recomendado para la mayoría de los proyectos. Elegilo si querés fluidez sin resignar los controles importantes; subí a Estricto si necesitás revisar más, o a Bypass si querés menos fricción.",

	"tier.bypass.label":   "Bypass",
	"tier.bypass.desc":    "Autonomía total opcional. El piso de seguridad de denegación sigue aplicando.",
	"tier.bypass.long":    "Bypass le da al agente autonomía total para actuar sin preguntar. El piso de seguridad de denegación sigue bloqueando las operaciones más peligrosas—esa capa no se desactiva—pero todo lo demás corre sin supervisión. Elegilo solo cuando confiás plenamente en el flujo y la fricción te frena más de lo que te protege. Usalo con precaución: es el nivel de menor control.",
	"tier.bypass.warning": "Bypass: modo autónomo — el piso de seguridad sigue aplicando (C-21)",

	// strategy.{targeted,restore}.{label,desc,long} — uninstall strategies.
	"strategy.targeted.label": "Dirigida",
	"strategy.targeted.desc":  "Revierte cada harness instalado de forma individual.",
	"strategy.targeted.long":  "Dirigida revierte cada harness instalado de forma individual, deshaciendo solo lo que esta herramienta cambió. No necesita un manifiesto de backup: recorre lo que se instaló y lo remueve pieza por pieza. Es la opción recomendada por defecto para la mayoría de las desinstalaciones. Elegila cuando querés sacar lo que instaló Active Stack sin tocar el resto de tu configuración.",

	"strategy.restore.label": "Restaurar desde backup",
	"strategy.restore.desc":  "Restaura el estado previo completo desde un manifiesto de backup.",
	"strategy.restore.long":  "Restaurar devuelve la máquina al estado exacto capturado en un manifiesto de backup previo a la instalación. En vez de deshacer harness por harness, reinstala la foto completa que se tomó antes de tocar tu config. Requiere que ese manifiesto esté disponible. Elegila cuando necesitás una reversión total en lugar de una eliminación puntual; si no tenés backup, usá Dirigida.",

	// backup.source.* — localized backup source labels.
	"backup.source.install": "instalación",
	"backup.source.sync":    "sincronización",
	"backup.source.upgrade": "actualización",
}

func init() {
	registerTable(LangES, tableES)
}
