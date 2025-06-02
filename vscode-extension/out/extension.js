"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || function (mod) {
    if (mod && mod.__esModule) return mod;
    var result = {};
    if (mod != null) for (var k in mod) if (k !== "default" && Object.prototype.hasOwnProperty.call(mod, k)) __createBinding(result, mod, k);
    __setModuleDefault(result, mod);
    return result;
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.deactivate = exports.activate = void 0;
const vscode = __importStar(require("vscode"));
const child_process_1 = require("child_process");
const util_1 = require("util");
const path = __importStar(require("path"));
const execAsync = (0, util_1.promisify)(child_process_1.exec);
function activate(context) {
    console.log('uroboro extension activating...');
    // Super minimal - just register commands, do nothing else
    context.subscriptions.push(vscode.commands.registerCommand('uroboro.capture', async () => {
        try {
            console.log('uroboro.capture called');
            const insight = await vscode.window.showInputBox({
                prompt: '💡 Enter your development insight:',
                placeHolder: 'Fixed auth timeout - cut query time from 3s to 200ms'
            });
            if (!insight?.trim()) {
                vscode.window.showWarningMessage('No insight entered');
                return;
            }
            // Find binary
            const binaryPaths = [
                '/home/qry/stuff/projects/uroboro/go/uroboro',
                path.join(__dirname, 'go-uroboro'),
                '/home/qry/stuff/projects/uroboro/uroboro-go'
            ];
            let goBinaryPath = '';
            for (const binPath of binaryPaths) {
                try {
                    if (require('fs').existsSync(binPath)) {
                        goBinaryPath = binPath;
                        break;
                    }
                }
                catch (e) {
                    console.log(`Error checking ${binPath}:`, e);
                }
            }
            if (!goBinaryPath) {
                vscode.window.showErrorMessage('uroboro binary not found');
                return;
            }
            console.log(`Using binary: ${goBinaryPath}`);
            // Run command with timeout
            const command = `${goBinaryPath} capture "${insight.trim()}"`;
            console.log(`Running: ${command}`);
            const { stdout, stderr } = await execAsync(command, { timeout: 10000 });
            if (stderr) {
                console.log('stderr:', stderr);
            }
            console.log('stdout:', stdout);
            vscode.window.showInformationMessage(`✅ Captured: ${insight.substring(0, 50)}...`);
        }
        catch (error) {
            console.error('Capture error:', error);
            vscode.window.showErrorMessage(`Capture failed: ${error.message}`);
        }
    }), vscode.commands.registerCommand('uroboro.status', async () => {
        try {
            console.log('uroboro.status called');
            vscode.window.showInformationMessage('Status command - simplified for debugging');
        }
        catch (error) {
            console.error('Status error:', error);
            vscode.window.showErrorMessage(`Status failed: ${error.message}`);
        }
    }));
    console.log('uroboro extension activated successfully');
}
exports.activate = activate;
function deactivate() {
    console.log('uroboro extension deactivated');
}
exports.deactivate = deactivate;
//# sourceMappingURL=extension.js.map