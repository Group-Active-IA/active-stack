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
	"tier.estricto.label":      "Estricto",
	"tier.estricto.codex.desc": "Codex: Solo lectura por defecto",
	"tier.estricto.codex.long": "Codex trabaja en solo lectura. Las acciones que requieren permisos adicionales necesitan aprobación explícita cuando la interfaz admite solicitarla. En las demás interfaces, esas acciones quedan bloqueadas.",
	"tier.estricto.desc":       "El agente debe pedir permiso para cada operación. Máxima fricción, máxima seguridad.",
	"tier.estricto.long":       "Estricto configura los permisos del agente para que pida confirmación antes de cada operación. Aplica el conjunto de reglas más restrictivo sobre el piso de seguridad: nada corre sin tu visto bueno. Es el nivel de mayor fricción y mayor seguridad. Elegilo para proyectos sensibles o si recién empezás y querés revisar todo antes de que pase; si te traba el ritmo, Balanceado afloja lo repetitivo.",

	"tier.balanceado.label": "Balanceado",
	"tier.balanceado.desc":  "Lista de permitidos curada para operaciones seguras y repetitivas. Punto de partida recomendado.",
	"tier.balanceado.long":  "Balanceado configura una lista de permitidos curada para que las operaciones seguras y repetitivas corran sin interrupciones, mientras las acciones más riesgosas siguen pidiendo confirmación. Es el punto de equilibrio entre fricción y autonomía sobre el piso de seguridad. Es el punto de partida recomendado para la mayoría de los proyectos. Elegilo si querés fluidez sin resignar los controles importantes; subí a Estricto si necesitás revisar más, o a Bypass si querés menos fricción.",

	"tier.bypass.label":   "Bypass",
	"tier.bypass.desc":    "Autonomía total opcional. El piso de seguridad de denegación sigue aplicando.",
	"tier.bypass.long":    "Bypass le da al agente autonomía total para actuar sin preguntar. El piso de seguridad de denegación sigue bloqueando las operaciones más peligrosas—esa capa no se desactiva—pero todo lo demás corre sin supervisión. Elegilo solo cuando confiás plenamente en el flujo y la fricción te frena más de lo que te protege. Usalo con precaución: es el nivel de menor control.",
	"tier.bypass.warning": "Bypass: modo autónomo — el piso de seguridad sigue aplicando",

	// claude-native-opsx: avisos de la entrega por capas de Claude.
	"claude.breaking.notice":           "BREAKING (Claude): el coordinador OPSX ya no vive en CLAUDE.md. Ahora es una skill que invocás con /sdd-orchestrator; las sesiones normales sólo conservan las reglas universales. Después de compactar la conversación, volvé a invocar /sdd-orchestrator. Tu contenido propio de CLAUDE.md se conservó.",
	"install.help.claude_user_content": "Nota sobre Claude: lo que escribas en tu propio CLAUDE.md llega a todas las sesiones y subagentes; Active Stack no lo puede filtrar. El instalador sólo administra sus secciones marcadas.",
	"update.breaking.nosync":           "Claude: esta instalación no tiene registro de selección, así que no se modificó CLAUDE.md. Cambio BREAKING: el coordinador OPSX ahora es la skill /sdd-orchestrator. Ejecutá `active-stack install` una vez con tu modo actual para migrar.",
	"update.project.legacy":            "El proyecto %s todavía tiene el bloque legado del coordinador en %s; no se reescribió. Migralo con: active-stack starter add <starter> --project %s",
	"install.project.manual_cleanup":   "Instalación de proyecto de Claude: para quitarla a mano, borrá %s y la sección active-stack:universal-rules de %s (el uninstall de proyecto todavía no está soportado).",
	"uninstall.flag.overwrite_edited":  "con --strategy restore, restaura CLAUDE.md y los archivos de la skill aunque se hayan editado después de la migración (antes guarda su contenido actual; --yes no lo implica)",

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

	// memory.* — "active-stack memory branch" ([EN] Fase 4).
	"memory.usage":                      "Uso: active-stack memory branch [--project <ruta>] [--json] [--lang en|es]\n     active-stack memory hook session-start|pre-read  (hook del agente: evento JSON por stdin)\n     active-stack memory init [--project <ruta>] [--name <nombre>] [--yes] [--json] [--lang en|es]\n     active-stack memory proxy [--project <ruta>] -- <comando de engram> [args...]  (proxy MCP que etiqueta los guardados)",
	"memory.proxy.no_command":           "memory proxy necesita el comando del servidor de Engram después de \"--\" (por ejemplo: active-stack memory proxy -- engram mcp).",
	"memory.label.mode":                "Modo de memoria",
	"memory.label.tagged":               "Etiqueta",
	"memory.label.branch":               "Rama",
	"memory.label.slug":                 "Slug",
	"memory.value.yes":                  "sí",
	"memory.value.no":                   "no",
	"memory.value.no_repo":              "(no es un repositorio Git)",
	"memory.value.none":                 "(ninguno)",
	"memory.warn.marker_invalid":        "Se ignora %s: no es JSON válido. El modo de memoria queda en off.",
	"memory.warn.marker_unknown_policy": "Se ignora %s: memory_policy %q desconocida. El modo de memoria queda en off.",
	"memory.error.unknown_subcommand":   "Subcomando de memory desconocido: %q.",
	"memory.error.no_directory":         "No existe el directorio: %s",
	"memory.error.git_missing":          "Git no está instalado o no está en el PATH.",
	"memory.error.empty_slug":           "La rama %q no tiene letras ni números para armar su etiqueta. Renombrá la rama.",
	"memory.error.failed":               "No se pudo leer el estado de Git: %v",
	"memory.hook.unknown":               "Hook de memoria desconocido: %q. Disponibles: session-start, pre-read. No se cambió nada.",
	"memory.hook.skipped":               "Se omitió el hook de memoria %s (%v). El agente sigue sin él.",
	"memory.init.project":               "Proyecto: %s",
	"memory.init.name":                  "Nombre del proyecto: %s (%s)",
	"memory.init.source.flag":           "de --name",
	"memory.init.source.remote":         "del remoto origin",
	"memory.init.source.directory":      "de la carpeta del checkout principal",
	"memory.init.file":                  "  %s: %s",
	"memory.init.action.created":        "creado",
	"memory.init.action.updated":        "actualizado",
	"memory.init.action.unchanged":      "sin cambios",
	"memory.init.backup":                "Backup de los archivos anteriores: %s",
	"memory.init.warn.dirty":            "Aviso: %s ya tenía cambios tuyos sin commitear. El commit propuesto los va a incluir.",
	"memory.init.warn.tracked":          "Aviso: otros archivos de .engram/ ya están versionados (%s). Las reglas de .gitignore no los des-versionan.",
	"memory.init.commit.header":         "Commit propuesto (NO se ejecutó: hacelo vos o pedilo). Corré los comandos desde %s:",
	"memory.init.commit.none":           "Nada para commitear: los archivos ya están versionados.",
	"memory.init.prompt":                "¿Usar %q como nombre del proyecto? [s/N] ",
	"memory.init.error.not_confirmed":   "El nombre del proyecto sería %q (%s). Confirmalo con --yes, o elegí otro con --name <nombre>. No se cambió nada.",
	"memory.init.error.declined":        "Cancelado: no se confirmó el nombre del proyecto. No se cambió nada.",
	"memory.init.error.conflict_name":   "Detenido: %s ya tiene project_name %q y active-stack escribiría %q. No se cambió nada. Si %[2]q es el nombre correcto, corré de nuevo con --name %[2]s; si no, editá el archivo a mano.",
	"memory.init.error.conflict_policy": "Detenido: %s tiene %s = %q y active-stack escribiría %q. No se cambió nada. Editá el archivo a mano y corré de nuevo.",
	"memory.init.error.ignored_dir":     "Detenido: .gitignore tiene la regla %q, que ignora toda la carpeta .engram/, y Git no puede reincluir archivos de una carpeta ignorada. Acotá esa regla (por ejemplo a .engram/*) y corré de nuevo. No se cambió nada.",
	"memory.init.error.still_ignored":   "Detenido: después de escribir, Git sigue ignorando %s por otra regla de ignorado (por ejemplo un ignore global, o una línea anterior de tu .gitignore como .*). Se devolvió todo a como estaba. Quitá esa regla y corré de nuevo.",
	"memory.init.error.not_ignored":     "Detenido: después de escribir, Git no ignoraría el resto de .engram/ (%s). Se devolvió todo a como estaba.",
	"memory.init.error.invalid_file":    "%s no se puede usar: %v. Corregilo y corré de nuevo. No se cambió nada.",
	"memory.init.error.home":            "No escribo en tu carpeta personal (%s). Corré el comando dentro de un repositorio de proyecto.",
	"memory.init.error.not_git_repo":    "Ese directorio no está dentro de un repositorio Git. No se cambió nada.",
	"memory.init.error.empty_name":      "No se pudo armar un nombre de proyecto desde el repositorio. Pasá uno con --name <nombre>.",
	"memory.init.error.failed":          "memory init falló: %v",
}

func init() {
	registerTable(LangES, tableES)
}
