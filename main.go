/*
Copyright © 2019 Ben Garrett <bengarrett77@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import "github.com/Defacto2/df2/lib/cmd"

/*
TODO's
- ansilove with no thumbs
- dz check for and update any metadata // titles, groups
-- also link any missing pouet ids
--> df2 fix demozoo (SLOW)


- strings to unicode runes for file contents and filenames
- uuid type enforcement
*/

func main() {
	cmd.Execute()
}

/*
   color.Info
   color.Note
   color.Light
   color.Error
   color.Danger
   color.Debug
   color.Notice
   color.Success
   color.Comment
   color.Primary
   color.Warning
   color.Question
   color.Secondary
*/
